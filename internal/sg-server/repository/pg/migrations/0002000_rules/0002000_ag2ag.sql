-- +goose Up
-- +goose StatementBegin

---------------------------------- AG2AG RULE -------------------------------------

drop table if exists sgroups.tbl_rule_registry cascade;
create table sgroups.tbl_rule_registry (
    uid        uuid not null primary key,
    name       sgroups.rname not null,
    ns         bigint,
    constraint registry_rule_uid_uq  unique (uid),
    constraint rule_registry_name_ns_uq unique (name, ns),
    constraint fk_rule_registry___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);
comment on table sgroups.tbl_rule_registry is 'registry for lookup rules of all kinds by primary keys';

drop type if exists sgroups.row_of__rule cascade;
create type sgroups.row_of__rule as (
  uid uuid,
  name text,
  namespace text
);

drop function if exists sgroups.sync_rule_via_registry(sgroups.sync_op, sgroups.row_of__rule) cascade;
create or replace function sgroups.sync_rule_via_registry(
  op sgroups.sync_op, d sgroups.row_of__rule
) returns void
as $$
declare
  nsID bigint;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;
  if op <> 'del' then
    raise exception 'unsupported op=% (use del)', op;
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

  delete from sgroups.tbl_rule_registry r
  where (
    not sgroups.is_empty_uuid((d).uid) and r.uid = (d).uid
  ) or (
    sgroups.is_empty_uuid((d).uid)
    and norm_name is not null
    and nsID is not null
    and r.name = norm_name::sgroups.rname
    and r.ns   = nsID
  );

  if not found then
    if not sgroups.is_empty_uuid((d).uid) then
      raise exception 'rule not found for uid=%', (d).uid
        using detail = 'SG0009',
              hint   = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    if norm_name is not null then
      raise exception 'rule not found for name=% namespace=%', norm_name, norm_ns
        using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    raise exception 'rule delete: either uid or name is required'
      using detail = 'SG0007',
            hint   = 'for delete pass existing uid OR (name AND namespace)';
  end if;
end;
$$ language plpgsql;

drop type if exists sgroups.port_entries cascade;
create type sgroups.port_entries as (
  description   text,
  comment text,
  ports sgroups.port_ranges
);
comment on type sgroups.port_entries is 'port entries type';

drop function if exists sgroups.rule_port_entries_no_overlap(sgroups.port_entries[]) cascade;
create or replace function sgroups.rule_port_entries_no_overlap(
  entries sgroups.port_entries[]
) returns boolean as $$
begin
  if entries is null or array_length(entries, 1) is null or array_length(entries, 1) <= 1 then
    return true;
  end if;
  return not coalesce(
    (select sgroups.i4mr_any_intersect((e).ports) from unnest(entries) e),
    false);
end;
$$ language plpgsql immutable;
comment on function sgroups.rule_port_entries_no_overlap
  is 'check that port ranges across entries within a single rule do not overlap';

drop function if exists sgroups.check_rule_port_entries_no_overlap_trg() cascade;
create or replace function sgroups.check_rule_port_entries_no_overlap_trg()
returns trigger as $$
begin
  if TG_OP = 'UPDATE' and NEW.entries is not distinct from OLD.entries then
    return NEW;
  end if;
  if not sgroups.rule_port_entries_no_overlap(NEW.entries) then
    raise exception 'rule "%" has overlapping port ranges across entries', NEW.name
      using detail = 'SG0011', hint = 'port ranges across entries of a rule must not overlap';
  end if;
  return NEW;
end;
$$ language plpgsql;


drop table if exists sgroups.tbl_ag2ag_rule cascade;
create table sgroups.tbl_ag2ag_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    agLocal bigint,
    agRemote bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto not null,
    ip_v sgroups.ip_family not null,
    entries sgroups.port_entries[],

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2ag_rule_uid_uq  unique (uid),
    constraint ag2ag_rule_name_uq unique (name, ns),
    constraint fk_ag2ag_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2ag_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2ag_rule___ag_remote
       foreign key(agRemote) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2ag_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists ag2ag_rule_labels_gin_idx
  on sgroups.tbl_ag2ag_rule using gin (labels);

create index if not exists ag2ag_rule_annotations_gin_idx
  on sgroups.tbl_ag2ag_rule using gin (annotations);

drop trigger if exists trg_ag2ag_rule_entries_no_overlap on sgroups.tbl_ag2ag_rule;
create trigger trg_ag2ag_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2ag_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_ag2ag_rule_immutable_fields on sgroups.tbl_ag2ag_rule;
create trigger trg_ag2ag_rule_immutable_fields
before update on sgroups.tbl_ag2ag_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2ag_rule_resource_version on sgroups.tbl_ag2ag_rule;
create trigger trg_ag2ag_rule_resource_version
before insert or update on sgroups.tbl_ag2ag_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2ag_rule_outbox_au on sgroups.tbl_ag2ag_rule;
create trigger trg_ag2ag_rule_outbox_au
after insert or update on sgroups.tbl_ag2ag_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2AgRule', 'sgroups.vu_ag2ag_rule');

drop trigger if exists trg_ag2ag_rule_outbox_bd on sgroups.tbl_ag2ag_rule;
create trigger trg_ag2ag_rule_outbox_bd
before delete on sgroups.tbl_ag2ag_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2AgRule', 'sgroups.vu_ag2ag_rule');


drop type if exists sgroups.endpoint cascade;
create type sgroups.endpoint as (
  res_type  sgroups.resource_type,
  name      text,
  namespace text,
  labels hstore
);
comment on type sgroups.endpoint is 'endpoint type';

drop view if exists sgroups.vu_ag2ag_rule cascade;
create or replace view sgroups.vu_ag2ag_rule as
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
  r.proto,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agl.name::text, ''),
    coalesce(agl_ns.name::text, ''),
    coalesce(agl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agr.name::text, ''),
    coalesce(agr_ns.name::text, ''),
    coalesce(agr.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as remote,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2ag_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns
left join sgroups.tbl_ag agr on agr.id = r.agRemote
left join sgroups.tbl_namespace agr_ns on agr_ns.id = agr.ns;


drop function if exists sgroups.list_ag2ag_rule() cascade;
create or replace function sgroups.list_ag2ag_rule()
returns setof sgroups.vu_ag2ag_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2ag_rule;
end;
$$ language plpgsql stable;


drop function if exists sgroups.resolve_ag2ag_rule_targets(text, sgroups.endpoint, sgroups.endpoint) cascade;
create or replace function sgroups.resolve_ag2ag_rule_targets(
  ruleNs text, agl sgroups.endpoint, agr sgroups.endpoint
)
returns table(agl_id bigint, agr_id bigint)
as $$
declare
  aglID bigint;
  aglName text;
  aglNs text;
  agrID bigint;
  agrName text;
  agrNs text;
begin
    if agl is null then
      raise exception 'sgroups.resolve_ag2ag_rule_targets: local address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing local address group reference';
    end if;
    if agr is null then
      raise exception 'sgroups.resolve_ag2ag_rule_targets: remote address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing remote address group reference';
    end if;

    aglName := nullif(btrim((agl).name), '');
    aglNs := nullif(btrim((agl).namespace), '');
    agrName := nullif(btrim((agr).name), '');
    agrNs := nullif(btrim((agr).namespace), '');

    if aglName is null or aglNs is null then
      raise exception 'sgroups.resolve_ag2ag_rule_targets: both local address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing local address group name and namespace';
    end if;
    if agrName is null or agrNs is null then
      raise exception 'sgroups.resolve_ag2ag_rule_targets: both remote address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing remote address group name and namespace';
    end if;

    if ruleNs <> aglNs then
      raise exception 'sgroups.resolve_ag2ag_rule_targets: rule namespace must match local address group namespace'
        using detail = 'SG0007', hint = 'pass rule namespace that matches local address group namespace';
    end if;

    select t.id
    from sgroups.tbl_ag t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = aglName::sgroups.rname
      and ns.name = aglNs::sgroups.rname
    into aglID;
    if aglID is null then
      raise exception 'local address group not found for name=% namespace=%', aglName, aglNs
        using detail = 'SG0009', hint = 'pass existing local address group name and namespace';
    end if;

    select t.id
    from sgroups.tbl_ag t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = agrName::sgroups.rname
      and ns.name = agrNs::sgroups.rname
    into agrID;
    if agrID is null then
      raise exception 'remote address group not found for name=% namespace=%', agrName, agrNs
        using detail = 'SG0009', hint = 'pass existing remote address group name and namespace';
    end if;

    agl_id := aglID;
    agr_id := agrID;
    return next;
end;
$$ language plpgsql;


drop type if exists sgroups.row_of__ag2ag_rule cascade;
create type sgroups.row_of__ag2ag_rule as (
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
  entries sgroups.port_entries[],
  proto sgroups.proto,
  ag_local sgroups.endpoint,
  ag_remote sgroups.endpoint
);


drop function if exists sgroups.sync_ag2ag_rule(sgroups.sync_op, sgroups.row_of__ag2ag_rule) cascade;
create or replace function sgroups.sync_ag2ag_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2ag_rule
)
returns setof sgroups.vu_ag2ag_rule
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
      raise exception 'sgroups.sync_ag2ag_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2ag_rule: namespace is required for op=ups'
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
          select * from sgroups.resolve_ag2ag_rule_targets(norm_ns, (d).ag_local, (d).ag_remote)
        )
        insert into sgroups.tbl_ag2ag_rule(
          uid, name, ns, agLocal, agRemote, display_name, labels, annotations, comment, description, action, traffic, proto, ip_v, entries
        )
        select
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          ids.agl_id,
          ids.agr_id,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).proto,
          (d).ip_v,
          (d).entries
        from ids
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2ag_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_ag2ag_rule_targets(norm_ns, (d).ag_local, (d).ag_remote)
      )
      update sgroups.tbl_ag2ag_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = ids.agl_id,
             agRemote = ids.agr_id,
             action = (d).action,
             traffic = (d).traffic,
             proto = (d).proto,
             ip_v = (d).ip_v,
             entries = (d).entries
        from ids
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2ag_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2ag_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2ag_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2ag_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2ag_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd