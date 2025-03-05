local k = import 'example.com/acme/jsonnet-libs/ksonnet-util/kausal.libsonnet',
      configMap = k.core.v1.configMap,
      container = k.core.v1.container,
      containerPort = k.core.v1.containerPort,
      deployment = k.apps.v1.deployment,
      job = k.batch.v1.job,
      clusterRole = k.rbac.v1.clusterRole,
      clusterRoleBinding = k.rbac.v1.clusterRoleBinding,
      policyRule = k.rbac.v1.policyRule,
      service = k.core.v1.service,
      serviceAccount = k.core.v1.serviceAccount,
      servicePort = k.core.v1.servicePort,
      subject = k.rbac.v1.subject,
      statefulSet = k.apps.v1.statefulSet;
local tracestore = import 'example.com/acme/tracestore/operations/jsonnet/microservices/tracestore.libsonnet';
local util = (import 'example.com/acme/jsonnet-libs/ksonnet-util/util.libsonnet').withK(k) {
  withNonRootSecurityContext(uid, fsGroup=null)::
    { spec+: { template+: { spec+: { securityContext: {
      fsGroup: if fsGroup == null then uid else fsGroup,
      runAsNonRoot: true,
      runAsUser: uid,
    } } } } },
};

tracestore {

  _common:: {
    licenseSecretName: 'get-license',
  },

  _images+:: {
    tracestore: 'acme/enterprise-traces:v1.3.0',
    kubectl: 'bitnami/kubectl',
  },

  _config+:: {
    // Enable to create the tokengen job, this job has to run once when installing GET to create a
    // token from the GET license. Once the token has been generated, this setting should be
    // disabled to not update the job anymore.
    create_tokengen_job: false,

    http_api_prefix: '/tracestore',
    otlp_port: 4317,
    distributor+: {
      replicas: 3,
      receivers: {
        otlp: {
          protocols: {
            grpc: {
              endpoint: '0.0.0.0:%d' % $._config.otlp_port,
            },
          },
        },
      },
    },
    ingester+: {
      replicas: 3,
      pvc_size: '50Gi',
      pvc_storage_class: 'fast',
    },
  },

  tracestore_config+:: {
    multitenancy_enabled: true,
    auth: { type: 'enterprise' },
    license: { path: '/etc/get-license/license.jwt' },
    admin_client: {
      storage: {
        type: $._config.backend,
        s3: {
          bucket_name: $._config.bucket + '-admin',
        },
        gcs: {
          bucket_name: $._config.bucket + '-admin',
        },
      },
    },
  },

  // create tracestore.yaml configuration file as ConfigMap
  tracestore_config_map:
    configMap.new('enterprise-traces')
    + configMap.withData({
      'tracestore.yaml': k.util.manifestYaml($.tracestore_config),
    }),

  // admin-api target
  admin_api_args:: {
    target: 'admin-api',
    'config.file': '/conf/tracestore.yaml',
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },
  admin_api_container::
    container.new(name='admin-api', image=$._images.tracestore)
    + container.withArgs(util.mapToFlags($.admin_api_args))
    + container.withPorts([containerPort.new('http-metrics', $._config.port)])
    + container.resources.withLimits({ cpu: '2', memory: '4Gi' })
    + container.resources.withRequests({ cpu: '500m', memory: '1Gi' })
    + tracestore.util.readinessProbe,
  admin_api_deployment:
    deployment.new(name='admin-api', replicas=1, containers=[$.admin_api_container])
    + deployment.spec.selector.withMatchLabelsMixin({ name: 'admin-api' })
    + deployment.spec.template.metadata.withLabelsMixin({ name: 'admin-api', [$._config.gossip_member_label]: 'true' })
    + deployment.spec.template.spec.withTerminationGracePeriodSeconds(15)
    + util.configVolumeMount($.tracestore_config_map.metadata.name, '/conf')
    + util.configVolumeMount($._config.overrides_configmap_name, '/overrides')
    + util.secretVolumeMount($._common.licenseSecretName, '/etc/get-license')
    + util.withNonRootSecurityContext(uid=10001),
  admin_api_service:
    util.serviceFor($.admin_api_deployment),

  // gateway target
  gateway_args:: {
    target: 'gateway',
    'config.file': '/conf/tracestore.yaml',
    'gateway.proxy.admin-api.url': 'http://admin-api:%s' % $._config.port,
    'gateway.proxy.compactor.url': 'http://compactor:%s' % $._config.port,
    'gateway.proxy.distributor.url': 'h2c://distributor:%s' % $._config.otlp_port,
    'gateway.proxy.ingester.url': 'http://ingester:%s' % $._config.port,
    'gateway.proxy.querier.url': 'http://querier:%s' % $._config.port,
    'gateway.proxy.query-frontend.url': 'http://query-frontend:%s' % $._config.port,
    'mem-ballast-size-mbs': $._config.ballast_size_mbs,
  },
  gateway_container::
    container.new(name='gateway', image=$._images.tracestore)
    + container.withArgs(util.mapToFlags($.gateway_args))
    + container.withPorts([containerPort.new('http-metrics', $._config.port)])
    + container.resources.withLimits({ cpu: '2', memory: '4Gi' })
    + container.resources.withRequests({ cpu: '500m', memory: '1Gi' })
    + tracestore.util.readinessProbe,
  gateway_deployment:
    deployment.new(name='gateway', replicas=1, containers=[$.gateway_container])
    + deployment.spec.template.spec.withTerminationGracePeriodSeconds(15)
    + util.configVolumeMount($.tracestore_config_map.metadata.name, '/conf')
    + util.configVolumeMount($._config.overrides_configmap_name, '/overrides')
    + util.secretVolumeMount($._common.licenseSecretName, '/etc/get-license')
    + util.withNonRootSecurityContext(uid=10001),
  gateway_service:
    util.serviceFor($.gateway_deployment),

  // tokengen target
  tokengen: {}
            + (if $._config.create_tokengen_job then {
                 local tokengen_args = {
                   target: 'tokengen',
                   'config.file': '/conf/tracestore.yaml',
                   'tokengen.token-file': '/shared/admin-token',
                 },
                 local tokengen_container =
                   container.new('tokengen', $._images.tracestore)
                   + container.withArgs(util.mapToFlags(tokengen_args))
                   + container.withVolumeMounts([
                     { mountPath: '/conf', name: $.tracestore_config_map.metadata.name },
                     { mountPath: '/shared', name: 'shared' },
                   ])
                   + container.resources.withLimits({ memory: '4Gi' })
                   + container.resources.withRequests({ cpu: '500m', memory: '500Mi' }),
                 local tokengen_create_secret_container =
                   container.new('create-secret', $._images.kubectl)
                   + container.withCommand([
                     '/bin/bash',
                     '-euc',
                     'kubectl create secret generic get-admin-token --from-file=token=/shared/admin-token --from-literal=acme-token="$(base64 <(echo :$(cat /shared/admin-token)))"',
                   ])
                   + container.withVolumeMounts([{ mountPath: '/shared', name: 'shared' }]),
                 tokengen_job:
                   job.new('tokengen')
                   + job.spec.withCompletions(1)
                   + job.spec.withParallelism(1)
                   + job.spec.template.spec.withContainers([tokengen_create_secret_container])
                   + job.spec.template.spec.withInitContainers([tokengen_container])
                   + job.spec.template.spec.withRestartPolicy('OnFailure')
                   + job.spec.template.spec.withServiceAccount('tokengen')
                   + job.spec.template.spec.withServiceAccountName('tokengen')
                   + job.spec.template.spec.withVolumes([
                     { name: $.tracestore_config_map.metadata.name, configMap: { name: $.tracestore_config_map.metadata.name } },
                     { name: 'shared', emptyDir: {} },
                   ])
                   + util.withNonRootSecurityContext(uid=10001),
                 tokengen_service_account:
                   serviceAccount.new('tokengen'),
                 tokengen_cluster_role:
                   clusterRole.new('tokengen')
                   + clusterRole.withRules([
                     policyRule.withApiGroups([''])
                     + policyRule.withResources(['secrets'])
                     + policyRule.withVerbs(['create']),
                   ]),
                 tokengen_cluster_role_binding:
                   clusterRoleBinding.new()
                   + clusterRoleBinding.metadata.withName('tokengen')
                   + clusterRoleBinding.roleRef.withApiGroup('rbac.authorization.k8s.io')
                   + clusterRoleBinding.roleRef.withKind('ClusterRole')
                   + clusterRoleBinding.roleRef.withName('tokengen')
                   + clusterRoleBinding.withSubjects([
                     subject.new()
                     + subject.withName('tokengen')
                     + subject.withKind('ServiceAccount')
                     + { namespace: $._config.namespace },
                   ]),
               } else {}),

  // upstream configuration overrides

  // distributor target
  tracestore_distributor_container+::
    container.withPortsMixin([
      containerPort.new('otlp-grpc', $._config.otlp_port),
    ]),

  // disable tracestore-vulture deployment until it supports basicauth
  tracestore_vulture_deployment:: {},
}
