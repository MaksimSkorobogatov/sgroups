-- +goose Up
-- +goose StatementBegin

drop type if exists sgroups.named_port cascade;
create type sgroups.named_port as (
  name sgroups.dname,
  port integer

);

drop type if exists sgroups.host_endpoints cascade;
create type sgroups.host_endpoints as (
    address inet,
    ports sgroups.named_port[]
);

alter table sgroups.tbl_host
  add column if not exists endpoints sgroups.host_endpoints;


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
  h.resource_version,
  h.endpoints
from sgroups.tbl_host as h
left join sgroups.tbl_namespace ns on ns.id = h.ns
;

drop type     if exists sgroups.row_of__host_ips cascade;
create type sgroups.row_of__host_ips as (
  uid       uuid,
  name      text,
  namespace text,
  ips       inet[],
  endpoints sgroups.host_endpoints
);

drop function if exists sgroups.sync_host_ipset(sgroups.sync_op, sgroups.row_of__host_ips) cascade;
create or replace function sgroups.sync_host_ipset(
  op sgroups.sync_op, d sgroups.row_of__host_ips
)
returns setof sgroups.vu_host
as $$
declare
  hostID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  if op <> 'ups' then
    raise exception 'sgroups.sync_host_ipset: unsupported op=% (use ups)', op;
  end if;

  norm_name := nullif(btrim((d).name), '');
  norm_ns := nullif(btrim((d).namespace), '');

  if sgroups.is_empty_uuid((d).uid) or norm_name is null or norm_ns is null then
      raise exception 'sgroups.sync_host_ipset: uid, name and namespace are required for updating host IPs'
        using detail = 'SG0007', hint = 'pass existing uid, name AND namespace';
  end if;

  select id
    into nsID
    from sgroups.tbl_namespace
   where name = norm_ns::sgroups.rname;
  if nsID is null then
    raise exception 'namespace not found for name=%', norm_ns
      using detail = 'SG0009', hint = 'pass existing namespace name';
  end if;

  select id
    into hostID
    from sgroups.tbl_host t
   where t.uid = (d).uid
     and t.name = norm_name::sgroups.rname
     and t.ns = nsID;
  if hostID is null then
    raise exception 'host not found for name=% namespace=% and uid=%', norm_name, norm_ns, (d).uid
      using detail = 'SG0009', hint = 'pass existing name, namespace and uid';
  end if;

  update sgroups.tbl_host t
     set ips       = coalesce((d).ips, '{}'::inet[]),
         endpoints = (d).endpoints
   where t.id = hostID;

  get diagnostics updated_count = row_count;
  if updated_count <> 1 then
    raise exception 'sgroups.sync_host_ipset: update mismatch (updated=%)', updated_count
      using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
  end if;

  return query
  select v.*
  from sgroups.vu_host v
  where v.uid = (d).uid;
  return;
end;
$$ language plpgsql strict;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd