-- +goose Up
-- +goose StatementBegin

drop table if exists sgroups.tbl_ag2svc_icmp_rule cascade;
create table sgroups.tbl_ag2svc_icmp_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    agLocal bigint,
    svcRemote bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    ip_v sgroups.ip_family not null,
    entries sgroups.icmp_entries[],

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2svc_icmp_rule_uid_uq  unique (uid),
    constraint ag2svc_icmp_rule_name_uq unique (name, ns),
    constraint fk_ag2svc_icmp_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2svc_icmp_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2svc_icmp_rule___svc_remote
       foreign key(svcRemote) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2svc_icmp_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists ag2svc_icmp_rule_labels_gin_idx
  on sgroups.tbl_ag2svc_icmp_rule using gin (labels);

create index if not exists ag2svc_icmp_rule_annotations_gin_idx
  on sgroups.tbl_ag2svc_icmp_rule using gin (annotations);

drop trigger if exists trg_ag2svc_icmp_rule_entries_no_overlap on sgroups.tbl_ag2svc_icmp_rule;
create trigger trg_ag2svc_icmp_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2svc_icmp_rule
for each row
execute function sgroups.check_rule_icmp_entries_no_overlap_trg();

drop trigger if exists trg_ag2svc_icmp_rule_immutable_fields on sgroups.tbl_ag2svc_icmp_rule;
create trigger trg_ag2svc_icmp_rule_immutable_fields
before update on sgroups.tbl_ag2svc_icmp_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2svc_icmp_rule_resource_version on sgroups.tbl_ag2svc_icmp_rule;
create trigger trg_ag2svc_icmp_rule_resource_version
before insert or update on sgroups.tbl_ag2svc_icmp_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2svc_icmp_rule_outbox_au on sgroups.tbl_ag2svc_icmp_rule;
create trigger trg_ag2svc_icmp_rule_outbox_au
after insert or update on sgroups.tbl_ag2svc_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2SvcIcmpRule', 'sgroups.vu_ag2svc_icmp_rule');

drop trigger if exists trg_ag2svc_icmp_rule_outbox_bd on sgroups.tbl_ag2svc_icmp_rule;
create trigger trg_ag2svc_icmp_rule_outbox_bd
before delete on sgroups.tbl_ag2svc_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2SvcIcmpRule', 'sgroups.vu_ag2svc_icmp_rule');


drop view if exists sgroups.vu_ag2svc_icmp_rule cascade;
create or replace view sgroups.vu_ag2svc_icmp_rule as
select
  r.uid,
  r.name,
  ns.name as namespace,
  r.display_name,
  r.comment,
  r.description,
  r.labels,
  r.annotations,
  r.action,
  r.traffic,
  r.ip_v,
  r.entries,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agl.name::text, ''),
    coalesce(agl_ns.name::text, ''),
    coalesce(agl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcr.name::text, ''),
    coalesce(svcr_ns.name::text, ''),
    coalesce(svcr.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as remote,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2svc_icmp_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns
left join sgroups.tbl_service svcr on svcr.id = r.svcRemote
left join sgroups.tbl_namespace svcr_ns on svcr_ns.id = svcr.ns;


drop function if exists sgroups.list_ag2svc_icmp_rule() cascade;
create or replace function sgroups.list_ag2svc_icmp_rule()
returns setof sgroups.vu_ag2svc_icmp_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2svc_icmp_rule;
end;
$$ language plpgsql stable;


drop type if exists sgroups.row_of__ag2svc_icmp_rule cascade;
create type sgroups.row_of__ag2svc_icmp_rule as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  action sgroups.policy_action,
  traffic sgroups.traffic,
  ip_v sgroups.ip_family,
  entries sgroups.icmp_entries[],
  ag_local sgroups.endpoint,
  svc_remote sgroups.endpoint
);


drop function if exists sgroups.sync_ag2svc_icmp_rule(sgroups.sync_op, sgroups.row_of__ag2svc_icmp_rule) cascade;
create or replace function sgroups.sync_ag2svc_icmp_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2svc_icmp_rule
)
returns setof sgroups.vu_ag2svc_icmp_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');
  norm_ns := nullif(btrim((d).namespace), '');

  if norm_name is not null and norm_ns is not null then
    select id
      into nsID
      from sgroups.tbl_namespace
     where name = norm_ns::sgroups.rname;
    if nsID is null then
      raise exception 'namespace not found for name=%', norm_ns
        using detail = 'SG0009', hint = 'pass existing namespace name';
    end if;
  end if;

  if op = 'del' then
    perform sgroups.sync_rule_via_registry(op, row((d).uid, (d).name, (d).namespace)::sgroups.row_of__rule);
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_ag2svc_icmp_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2svc_icmp_rule: namespace is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing namespace for both insert (uid=null) and update (uid!=null)';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      rl_uid := gen_random_uuid();
      begin
        insert into sgroups.tbl_rule_registry(uid, name, ns)
        values (rl_uid, norm_name::sgroups.rname, nsID);
      exception
        when unique_violation then
          raise exception 'rule already exists for name=% namespace=% (cross-type uniqueness)',
                norm_name, norm_ns
            using detail = 'SG0008',
                  hint   = 'rule name must be unique across all rule types within namespace';
      end;

      begin
        with ids as (
          select * from sgroups.resolve_ag2svc_rule_targets(norm_ns, (d).ag_local, (d).svc_remote)
        )
        insert into sgroups.tbl_ag2svc_icmp_rule(
          uid, name, ns, agLocal, svcRemote, display_name, labels, annotations,
          comment, description, action, traffic, ip_v, entries
        )
        select
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          ids.agl_id,
          ids.svcr_id,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries
        from ids
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'ag2svc icmp rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2svc_icmp_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_ag2svc_rule_targets(norm_ns, (d).ag_local, (d).svc_remote)
      )
      update sgroups.tbl_ag2svc_icmp_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = ids.agl_id,
             svcRemote = ids.svcr_id,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries
        from ids
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'ag2svc icmp rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2svc_icmp_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2svc_icmp_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2svc_icmp_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2svc_icmp_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2svc_icmp_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
