{
  folderID(folder)::
    local lower = std.asciiLower(folder);
    local underscore = std.strReplace(lower, '_', '-');
    local space = std.strReplace(underscore, ' ', '-');
    space,

  local folderID(folder) = self.folderID(folder),

  // add a new empty folder
  // It's super common for the dashboards in a single folder to be too
  // large to fit inside a single Kubernetes ConfigMap. In that case,
  // dashboards will be sharded into multiple ConfigMaps
  addFolder(name):: {
    acmeDashboardFolders+:: {
      [name]: {
        id: folderID(name),
        name: name,
        dashboards: {},
      },
    },
  },

  // add a new dashboard, creating the folder if necessary
  addDashboard(name, dashboard, folder=''):: {
    acmeDashboardFolders+:: {
      [folder]+: {
        id: folderID(folder),
        name: folder,
        dashboards+: {
          [name]+: dashboard,
        },
      },
    },
  },

  addMixinDashboards(mixins, mixinProto={}):: {
    local acmeDashboards = super.acmeDashboards,
    acmeDashboardFolders+:: std.foldr(
      function(name, acc)
        acc
        + (
          if std.objectHasAll(mixins[name], 'acmeDashboards')
             && std.length(mixins[name].acmeDashboards) > 0
          then
            local key = (
              if std.objectHasAll(mixins[name], 'acmeDashboardFolder')
              then $.folderID(mixins[name].acmeDashboardFolder)
              else 'general'
            );
            {
              [key]+: {
                dashboards+: (mixins[name] + mixinProto).acmeDashboards,
                name:
                  if std.objectHasAll(mixins[name], 'acmeDashboardFolder')
                  then mixins[name].acmeDashboardFolder
                  else '',
                id: $.folderID(self.name),
              },
            }
          else {}
        ),
      std.objectFields(mixins),
      {}
    ) + {
      general+: {
        dashboards+: acmeDashboards + mixinProto,
        name: '',
        id: '',
      },
    },
  },

  acmeDashboardFolders+:: {},
}
