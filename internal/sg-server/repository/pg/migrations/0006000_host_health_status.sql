-- +goose Up
-- +goose StatementBegin

alter table sgroups.tbl_host
  add column if not exists health_status boolean;

create or replace view sgroups.vu_host as
select
  h.uid, h.name, ns.name as namespace,
  h.display_name, h.comment, h.description,
  h.labels, h.annotations,
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
    (select array_agg(hb.ag order by hb.ag)
     from sgroups.tbl_host_binding hb where hb.host = h.id)
  ) as refs,
  h.creation_timestamp,
  h.resource_version,
  h.endpoints,
  h.health_status
from sgroups.tbl_host as h
left join sgroups.tbl_namespace ns on ns.id = h.ns;

drop type if exists sgroups.row_of__host_health cascade;
create type sgroups.row_of__host_health as (
  uid            uuid,
  name           text,
  namespace      text,
  health_status  boolean
);

drop function if exists sgroups.sync_host_health_status(sgroups.sync_op, sgroups.row_of__host_health) cascade;
create or replace function sgroups.sync_host_health_status(
  op sgroups.sync_op, d sgroups.row_of__host_health
)
returns setof sgroups.vu_host
as $$
declare
  hostID bigint;
  nsID bigint;
  norm_name text;
  norm_ns text;
begin
  if d is null then return; end if;
  if op <> 'ups' then
    raise exception 'sgroups.sync_host_health_status: unsupported op=%', op;
  end if;

  norm_name := nullif(btrim((d).name), '');
  norm_ns := nullif(btrim((d).namespace), '');

  if sgroups.is_empty_uuid((d).uid) or norm_name is null or norm_ns is null then
    raise exception 'uid, name and namespace are required';
  end if;

  select id into nsID from sgroups.tbl_namespace where name = norm_ns::sgroups.rname;
  if nsID is null then
    raise exception 'namespace not found for name=%', norm_ns;
  end if;

  select id into hostID from sgroups.tbl_host t
   where t.uid = (d).uid and t.name = norm_name::sgroups.rname and t.ns = nsID;
  if hostID is null then
    raise exception 'host not found for name=% namespace=% uid=%', norm_name, norm_ns, (d).uid;
  end if;

  update sgroups.tbl_host t
     set health_status = (d).health_status
   where t.id = hostID;

  return query select v.* from sgroups.vu_host v where v.uid = (d).uid;
end;
$$ language plpgsql strict;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

create or replace view sgroups.vu_host as
select
  h.uid, h.name, ns.name as namespace,
  h.display_name, h.comment, h.description,
  h.labels, h.annotations,
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
    (select array_agg(hb.ag order by hb.ag)
     from sgroups.tbl_host_binding hb where hb.host = h.id)
  ) as refs,
  h.creation_timestamp,
  h.resource_version,
  h.endpoints
from sgroups.tbl_host as h
left join sgroups.tbl_namespace ns on ns.id = h.ns;

drop function if exists sgroups.sync_host_health_status(sgroups.sync_op, sgroups.row_of__host_health) cascade;
drop type if exists sgroups.row_of__host_health cascade;
alter table sgroups.tbl_host drop column if exists health_status;

-- +goose StatementEnd
