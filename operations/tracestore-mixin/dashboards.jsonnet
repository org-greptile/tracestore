local mixin = import 'mixin.libsonnet';

{
  [name]: std.manifestJsonEx(mixin.acmeDashboards[name], ' ')
  for name in std.objectFields(mixin.acmeDashboards)
}
