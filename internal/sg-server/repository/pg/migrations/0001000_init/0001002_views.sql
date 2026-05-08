-- +goose Up
-- +goose StatementBegin

---------------------------------- RESOURCE VIEWERS AND HELPERS -------------------------------------

drop view if exists sgroups.vu_namespace cascade;
create or replace view sgroups.vu_namespace as
select
  uid, name, display_name, comment, description,
  labels, annotations, creation_timestamp, resource_version
from sgroups.tbl_namespace;


drop function if exists sgroups.get_ag_refs(bigint[]) cascade;
create or replace function sgroups.get_ag_refs(ag_ids bigint[])
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if ag_ids is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        ag.name::text,
        ns.name::text,
        'AddressGroup'::sgroups.resource_type
      )::sgroups.resource_ref
      order by ag.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag ag
  left join sgroups.tbl_namespace ns on ns.id = ag.ns
  where ag.id = any(ag_ids);

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag_refs(bigint[])
  is 'Returns AddressGroup refs by ids';

drop function if exists sgroups.get_ag_refs(bigint) cascade;
create or replace function sgroups.get_ag_refs(ag_id bigint)
returns sgroups.resource_ref[]
as $$
begin
  if ag_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  return sgroups.get_ag_refs(array[ag_id]::bigint[]);
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag_refs(bigint)
  is 'Returns AddressGroup ref';


drop view if exists sgroups.vu_network cascade;
create or replace view sgroups.vu_network as
select
  nw.uid,
  nw.name,
  ns.name as namespace,
  nw.display_name,
  nw.comment,
  nw.description,
  nw.labels,
  nw.annotations,
  nw.network,
  sgroups.get_ag_refs(
    (
      select array_agg(nb.ag order by nb.ag)
      from sgroups.tbl_network_binding nb
      where nb.network = nw.id
    )
  ) as refs,
  nw.creation_timestamp,
  nw.resource_version
from sgroups.tbl_network as nw
left join sgroups.tbl_namespace ns on ns.id = nw.ns;

drop view if exists sgroups.vu_host cascade;
create or replace view sgroups.vu_host as
select
  h.uid,
  h.name,
  ns.name as namespace,
  h.display_name,
  h.comment,
  h.description,
  h.labels,
  h.annotations,
  h.ips,
  row(
    coalesce((h.meta_info).host_name, ''::sgroups.dname),
    coalesce((h.meta_info).os, ''::sgroups.dname),
    coalesce((h.meta_info).platform, ''::sgroups.dname),
    coalesce((h.meta_info).platform_family, ''::sgroups.dname),
    coalesce((h.meta_info).platform_version, ''::sgroups.dname),
    coalesce((h.meta_info).kernel_version, ''::sgroups.dname)
  )::sgroups.host_info as meta_info,
  sgroups.get_ag_refs(
    (
      select array_agg(hb.ag order by hb.ag)
      from sgroups.tbl_host_binding hb
      where hb.host = h.id
    )
  ) as refs,
  h.creation_timestamp,
  h.resource_version
from sgroups.tbl_host as h
left join sgroups.tbl_namespace ns on ns.id = h.ns
;


drop function if exists sgroups.get_host_refs(bigint) cascade;
create or replace function sgroups.get_host_refs(ag_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if ag_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        h.name::text,
        ns.name::text,
        'Host'::sgroups.resource_type
      )::sgroups.resource_ref
      order by h.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_host h
  join sgroups.tbl_namespace ns on ns.id = h.ns
  join sgroups.tbl_host_binding hb on hb.host = h.id
  where hb.ag = ag_id;

  return r;
end;
$$ language plpgsql stable;


drop function if exists sgroups.get_network_refs(bigint) cascade;
create or replace function sgroups.get_network_refs(ag_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if ag_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        nw.name::text,
        ns.name::text,
        'Network'::sgroups.resource_type
      )::sgroups.resource_ref
      order by nw.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_network nw
  join sgroups.tbl_namespace ns on ns.id = nw.ns
  join sgroups.tbl_network_binding nb on nb.network = nw.id
  where nb.ag = ag_id;

  return r;
end;
$$ language plpgsql stable;


drop function if exists sgroups.get_service_refs(bigint) cascade;
create or replace function sgroups.get_service_refs(ag_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if ag_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        svc.name::text,
        ns.name::text,
        'Service'::sgroups.resource_type
      )::sgroups.resource_ref
      order by svc.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_service svc
  join sgroups.tbl_namespace ns on ns.id = svc.ns
  join sgroups.tbl_service_binding sb on sb.service = svc.id
  where sb.ag = ag_id;

  return r;
end;
$$ language plpgsql stable;


drop view if exists sgroups.vu_ag cascade;
create or replace view sgroups.vu_ag as
select
  ag.uid,
  ag.name,
  (select name from sgroups.tbl_namespace where id = ag.ns) as namespace,
  ag.display_name,
  ag.comment,
  ag.description,
  ag.labels,
  ag.annotations,
  ag.default_action,
  ag.logs,
  ag.trace,
  sgroups.get_host_refs(ag.id) || sgroups.get_network_refs(ag.id) || sgroups.get_service_refs(ag.id) as refs,
  ag.creation_timestamp,
  ag.resource_version
from sgroups.tbl_ag as ag;

drop view if exists sgroups.vu_service cascade;
create or replace view sgroups.vu_service as
select
  svc.uid,
  svc.name,
  ns.name as namespace,
  svc.display_name,
  svc.comment,
  svc.description,
  svc.labels,
  svc.annotations,
  svc.transports,
  sgroups.get_ag_refs(
    (
      select array_agg(sb.ag order by sb.ag)
      from sgroups.tbl_service_binding sb
      where sb.service = svc.id
    )
  ) as refs,
  svc.creation_timestamp,
  svc.resource_version
from sgroups.tbl_service as svc
left join sgroups.tbl_namespace ns on ns.id = svc.ns;

drop view if exists sgroups.vu_host_binding cascade;
create or replace view sgroups.vu_host_binding as
select
  hb.uid,
  hb.name,
  hb_ns.name as namespace,
  hb.display_name,
  hb.comment,
  hb.description,
  hb.labels,
  hb.annotations,
  row(
    coalesce(ag.name::text, ''),
    coalesce(ag_ns.name::text, '')
  )::sgroups.resource_id as ag,
  row(
    coalesce(h.name::text, ''),
    coalesce(h_ns.name::text, '')
  )::sgroups.resource_id as host,

  hb.creation_timestamp,
  hb.resource_version
from sgroups.tbl_host_binding as hb
left join sgroups.tbl_namespace hb_ns on hb_ns.id = hb.ns
left join sgroups.tbl_ag ag on ag.id = hb.ag
left join sgroups.tbl_namespace ag_ns on ag_ns.id = ag.ns
left join sgroups.tbl_host h on h.id = hb.host
left join sgroups.tbl_namespace h_ns on h_ns.id = h.ns;


drop view if exists sgroups.vu_network_binding cascade;
create or replace view sgroups.vu_network_binding as
select
  nb.uid,
  nb.name,
  nb_ns.name as namespace,
  nb.display_name,
  nb.comment,
  nb.description,
  nb.labels,
  nb.annotations,
  row(
    coalesce(ag.name::text, ''),
    coalesce(ag_ns.name::text, '')
  )::sgroups.resource_id as ag,
  row(
    coalesce(nw.name::text, ''),
    coalesce(nw_ns.name::text, '')
  )::sgroups.resource_id as network,

  nb.creation_timestamp,
  nb.resource_version
from sgroups.tbl_network_binding as nb
left join sgroups.tbl_namespace nb_ns on nb_ns.id = nb.ns
left join sgroups.tbl_ag ag on ag.id = nb.ag
left join sgroups.tbl_namespace ag_ns on ag_ns.id = ag.ns
left join sgroups.tbl_network nw on nw.id = nb.network
left join sgroups.tbl_namespace nw_ns on nw_ns.id = nw.ns;


drop view if exists sgroups.vu_service_binding cascade;
create or replace view sgroups.vu_service_binding as
select
  sb.uid,
  sb.name,
  sb_ns.name as namespace,
  sb.display_name,
  sb.comment,
  sb.description,
  sb.labels,
  sb.annotations,
  row(
    coalesce(ag.name::text, ''),
    coalesce(ag_ns.name::text, '')
  )::sgroups.resource_id as ag,
  row(
    coalesce(svc.name::text, ''),
    coalesce(svc_ns.name::text, '')
  )::sgroups.resource_id as service,

  sb.creation_timestamp,
  sb.resource_version
from sgroups.tbl_service_binding as sb
left join sgroups.tbl_namespace sb_ns on sb_ns.id = sb.ns
left join sgroups.tbl_ag ag on ag.id = sb.ag
left join sgroups.tbl_namespace ag_ns on ag_ns.id = ag.ns
left join sgroups.tbl_service svc on svc.id = sb.service
left join sgroups.tbl_namespace svc_ns on svc_ns.id = svc.ns;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd