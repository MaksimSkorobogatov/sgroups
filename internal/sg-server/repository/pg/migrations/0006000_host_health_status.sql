-- +goose Up
-- +goose StatementBegin

alter table sgroups.tbl_host
  add column if not exists healthy boolean not null default false;

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
  h.healthy
from sgroups.tbl_host as h
left join sgroups.tbl_namespace ns on ns.id = h.ns;

drop type if exists sgroups.row_of__host_health cascade;
create type sgroups.row_of__host_health as (
  uid       uuid,
  name      text,
  namespace text,
  healthy   boolean
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
    raise exception 'sgroups.sync_host_health_status: uid, name and namespace are required'
      using detail = 'SG0007', hint = 'pass existing uid, name AND namespace';
  end if;

  select id into nsID from sgroups.tbl_namespace where name = norm_ns::sgroups.rname;
  if nsID is null then
    raise exception 'namespace not found for name=%', norm_ns
      using detail = 'SG0009', hint = 'pass existing namespace name';
  end if;

  select id into hostID from sgroups.tbl_host t
   where t.uid = (d).uid and t.name = norm_name::sgroups.rname and t.ns = nsID;
  if hostID is null then
    raise exception 'host not found for name=% namespace=% uid=%', norm_name, norm_ns, (d).uid
      using detail = 'SG0009', hint = 'pass existing name, namespace and uid';
  end if;

  update sgroups.tbl_host t
     set healthy = (d).healthy
   where t.id = hostID
     and t.healthy is distinct from (d).healthy;

  return query select v.* from sgroups.vu_host v where v.uid = (d).uid;
end;
$$ language plpgsql strict;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop function if exists sgroups.sync_host_health_status(sgroups.sync_op, sgroups.row_of__host_health) cascade;
drop function if exists sgroups.sync_host_ipset(sgroups.sync_op, sgroups.row_of__host_ips) cascade;
drop function if exists sgroups.sync_host_info(sgroups.sync_op, sgroups.row_of__host_info) cascade;
drop function if exists sgroups.sync_hosts(sgroups.sync_op, sgroups.row_of__host) cascade;
drop function if exists sgroups.list_hosts(sgroups.res_selector[]) cascade;

drop view if exists sgroups.vu_host;

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

drop type if exists sgroups.row_of__host_health cascade;
alter table sgroups.tbl_host drop column if exists healthy;

create or replace function sgroups.sync_hosts(
  op sgroups.sync_op, d sgroups.row_of__host
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

  norm_name := nullif(btrim((d).name), '');
  norm_ns := nullif(btrim((d).namespace), '');

  if sgroups.is_empty_uuid((d).uid) or norm_name is null or norm_ns is null then
    raise exception 'sgroups.sync_hosts: uid, name and namespace are required'
      using detail = 'SG0007', hint = 'pass existing uid, name AND namespace';
  end if;

  select id into nsID from sgroups.tbl_namespace where name = norm_ns::sgroups.rname;
  if nsID is null then
    raise exception 'namespace not found for name=%', norm_ns
      using detail = 'SG0009', hint = 'pass existing namespace name';
  end if;

  if op = 'ups' then
    insert into sgroups.tbl_host (uid, name, ns, labels, annotations, comment, description, display_name)
    values (
      (d).uid,
      norm_name::sgroups.rname,
      nsID,
      (d).labels,
      (d).annotations,
      (d).comment,
      (d).description,
      (d).display_name
    )
    on conflict (uid) do update set
      name        = excluded.name,
      ns          = excluded.ns,
      labels      = excluded.labels,
      annotations = excluded.annotations,
      comment     = excluded.comment,
      description = excluded.description,
      display_name = excluded.display_name
    where sgroups.tbl_host.uid = excluded.uid
    returning id into hostID;

    if hostID is null then
      select id into hostID from sgroups.tbl_host where uid = (d).uid;
    end if;
  elsif op = 'del' then
    select id into hostID from sgroups.tbl_host t
     where t.uid = (d).uid and t.name = norm_name::sgroups.rname and t.ns = nsID;
    if hostID is not null then
      delete from sgroups.tbl_host_binding hb where hb.host = hostID;
      delete from sgroups.tbl_host t where t.id = hostID;
    end if;
  else
    raise exception 'sgroups.sync_hosts: unsupported op=%', op;
  end if;

  return query select v.* from sgroups.vu_host v where v.uid = (d).uid;
end;
$$ language plpgsql strict;

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

create or replace function sgroups.sync_host_info(
  op sgroups.sync_op, d sgroups.row_of__host_info
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
    raise exception 'sgroups.sync_host_info: unsupported op=%', op;
  end if;

  norm_name := nullif(btrim((d).name), '');
  norm_ns := nullif(btrim((d).namespace), '');

  if sgroups.is_empty_uuid((d).uid) or norm_name is null or norm_ns is null then
    raise exception 'sgroups.sync_host_info: uid, name and namespace are required'
      using detail = 'SG0007', hint = 'pass existing uid, name AND namespace';
  end if;

  select id into nsID from sgroups.tbl_namespace where name = norm_ns::sgroups.rname;
  if nsID is null then
    raise exception 'namespace not found for name=%', norm_ns
      using detail = 'SG0009', hint = 'pass existing namespace name';
  end if;

  select id into hostID from sgroups.tbl_host t
   where t.uid = (d).uid and t.name = norm_name::sgroups.rname and t.ns = nsID;
  if hostID is null then
    raise exception 'host not found for name=% namespace=% uid=%', norm_name, norm_ns, (d).uid
      using detail = 'SG0009', hint = 'pass existing name, namespace and uid';
  end if;

  update sgroups.tbl_host t
     set meta_info = (d).meta_info
   where t.id = hostID;

  return query select v.* from sgroups.vu_host v where v.uid = (d).uid;
end;
$$ language plpgsql strict;

create or replace function sgroups.list_hosts(selectors sgroups.res_selector[] default null)
returns setof sgroups.vu_host
as $$
declare
  has_complex   boolean;
  has_match_all boolean;
begin
  if selectors is null or cardinality(selectors) = 0 then
    return query select * from sgroups.vu_host;
    return;
  end if;

  select
    bool_or(
      (s.label_selector is not null
        and cardinality(akeys(s.label_selector)) > 0)
      or (
        (s.field_selector).refs is not null
        and cardinality((s.field_selector).refs) > 0
      )
    ),
    bool_or(
      nullif(btrim((s.field_selector).name), '') is null
      and nullif(btrim((s.field_selector).namespace), '') is null
      and (s.label_selector is null
           or cardinality(akeys(s.label_selector)) = 0)
      and (
        (s.field_selector).refs is null
        or cardinality((s.field_selector).refs) = 0
      )
    )
  into has_complex, has_match_all
  from unnest(selectors) s;

  if has_match_all then
    return query select * from sgroups.vu_host;
    return;
  end if;

  if has_complex then
    return query
    select v.*
    from sgroups.vu_host v
    where sgroups.match_res_selectors(
      sgroups.mk_res_selector(v.name, v.namespace, v.refs, v.labels),
      selectors
    );
    return;
  end if;

  return query
  select v.*
  from sgroups.vu_host v
  where exists (
    select 1
    from unnest(selectors) s
    where (
        nullif(btrim((s.field_selector).name), '') is null
        or nullif(btrim((s.field_selector).name), '') = v.name
      )
      and (
        nullif(btrim((s.field_selector).namespace), '') is null
        or nullif(btrim((s.field_selector).namespace), '') = v.namespace
      )
  );
end;
$$ language plpgsql stable;

-- +goose StatementEnd
