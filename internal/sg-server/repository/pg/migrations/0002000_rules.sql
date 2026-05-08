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

---------------------------------- AG2AG ICMP RULE -------------------------------------


drop type if exists sgroups.icmp_entries cascade;
create type sgroups.icmp_entries as (
  description   text,
  comment text,
  types sgroups.icmp_types
);
comment on type sgroups.icmp_entries is 'icmp entries type';

drop function if exists sgroups.rule_icmp_entries_no_overlap(sgroups.icmp_entries[]) cascade;
create or replace function sgroups.rule_icmp_entries_no_overlap(
  entries sgroups.icmp_entries[]
) returns boolean as $$
begin
  if entries is null or array_length(entries, 1) is null or array_length(entries, 1) <= 1 then
    return true;
  end if;
  return not exists (
    select 1
    from unnest(entries) as e1,
         unnest(entries) as e2
    where e1 is distinct from e2
      and (e1).types && (e2).types
  );
end;
$$ language plpgsql immutable;
comment on function sgroups.rule_icmp_entries_no_overlap
  is 'check that ICMP types across entries within a single rule do not overlap';

drop function if exists sgroups.check_rule_icmp_entries_no_overlap_trg() cascade;
create or replace function sgroups.check_rule_icmp_entries_no_overlap_trg()
returns trigger as $$
begin
  if TG_OP = 'UPDATE' and NEW.entries is not distinct from OLD.entries then
    return NEW;
  end if;
  if not sgroups.rule_icmp_entries_no_overlap(NEW.entries) then
    raise exception 'rule "%" has overlapping ICMP types across entries', NEW.name
      using detail = 'SG0011', hint = 'ICMP types across entries of a rule must not overlap';
  end if;
  return NEW;
end;
$$ language plpgsql;


drop table if exists sgroups.tbl_ag2ag_icmp_rule cascade;
create table sgroups.tbl_ag2ag_icmp_rule (
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
    ip_v sgroups.ip_family not null,
    entries sgroups.icmp_entries[],

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2ag_icmp_rule_uid_uq  unique (uid),
    constraint ag2ag_icmp_rule_name_uq unique (name, ns),
    constraint fk_ag2ag_icmp_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2ag_icmp_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2ag_icmp_rule___ag_remote
       foreign key(agRemote) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2ag_icmp_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists ag2ag_icmp_rule_labels_gin_idx
  on sgroups.tbl_ag2ag_icmp_rule using gin (labels);

create index if not exists ag2ag_icmp_rule_annotations_gin_idx
  on sgroups.tbl_ag2ag_icmp_rule using gin (annotations);

drop trigger if exists trg_ag2ag_icmp_rule_entries_no_overlap on sgroups.tbl_ag2ag_icmp_rule;
create trigger trg_ag2ag_icmp_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2ag_icmp_rule
for each row
execute function sgroups.check_rule_icmp_entries_no_overlap_trg();

drop trigger if exists trg_ag2ag_icmp_rule_immutable_fields on sgroups.tbl_ag2ag_icmp_rule;
create trigger trg_ag2ag_icmp_rule_immutable_fields
before update on sgroups.tbl_ag2ag_icmp_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2ag_icmp_rule_resource_version on sgroups.tbl_ag2ag_icmp_rule;
create trigger trg_ag2ag_icmp_rule_resource_version
before insert or update on sgroups.tbl_ag2ag_icmp_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2ag_icmp_rule_outbox_au on sgroups.tbl_ag2ag_icmp_rule;
create trigger trg_ag2ag_icmp_rule_outbox_au
after insert or update on sgroups.tbl_ag2ag_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2AgIcmpRule', 'sgroups.vu_ag2ag_icmp_rule');

drop trigger if exists trg_ag2ag_icmp_rule_outbox_bd on sgroups.tbl_ag2ag_icmp_rule;
create trigger trg_ag2ag_icmp_rule_outbox_bd
before delete on sgroups.tbl_ag2ag_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2AgIcmpRule', 'sgroups.vu_ag2ag_icmp_rule');


drop view if exists sgroups.vu_ag2ag_icmp_rule cascade;
create or replace view sgroups.vu_ag2ag_icmp_rule as
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
    'AddressGroup'::sgroups.resource_type,
    coalesce(agr.name::text, ''),
    coalesce(agr_ns.name::text, ''),
    coalesce(agr.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as remote,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2ag_icmp_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns
left join sgroups.tbl_ag agr on agr.id = r.agRemote
left join sgroups.tbl_namespace agr_ns on agr_ns.id = agr.ns;


drop function if exists sgroups.list_ag2ag_icmp_rule() cascade;
create or replace function sgroups.list_ag2ag_icmp_rule()
returns setof sgroups.vu_ag2ag_icmp_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2ag_icmp_rule;
end;
$$ language plpgsql stable;



drop type if exists sgroups.row_of__ag2ag_icmp_rule cascade;
create type sgroups.row_of__ag2ag_icmp_rule as (
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
  ag_remote sgroups.endpoint
);


drop function if exists sgroups.sync_ag2ag_icmp_rule(sgroups.sync_op, sgroups.row_of__ag2ag_icmp_rule) cascade;
create or replace function sgroups.sync_ag2ag_icmp_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2ag_icmp_rule
)
returns setof sgroups.vu_ag2ag_icmp_rule
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
      raise exception 'sgroups.sync_ag2ag_icmp_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2ag_icmp_rule: namespace is required for op=ups'
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
        insert into sgroups.tbl_ag2ag_icmp_rule(
          uid, name, ns, agLocal, agRemote, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries
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
          (d).ip_v,
          (d).entries
        from ids
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'ag2ag icmp rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2ag_icmp_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_ag2ag_rule_targets(norm_ns, (d).ag_local, (d).ag_remote)
      )
      update sgroups.tbl_ag2ag_icmp_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = ids.agl_id,
             agRemote = ids.agr_id,
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
        raise exception 'ag2ag icmp rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2ag_icmp_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2ag_icmp_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2ag_icmp_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2ag_icmp_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2ag_icmp_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- AG2ICMP RULE -------------------------------------


drop table if exists sgroups.tbl_ag2icmp_rule cascade;
create table sgroups.tbl_ag2icmp_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    agLocal bigint,
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

    constraint ag2icmp_rule_uid_uq  unique (uid),
    constraint ag2icmp_rule_name_uq unique (name, ns),
    constraint fk_ag2icmp_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2icmp_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2icmp_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists ag2icmp_rule_labels_gin_idx
  on sgroups.tbl_ag2icmp_rule using gin (labels);

create index if not exists ag2icmp_rule_annotations_gin_idx
  on sgroups.tbl_ag2icmp_rule using gin (annotations);

drop trigger if exists trg_ag2icmp_rule_entries_no_overlap on sgroups.tbl_ag2icmp_rule;
create trigger trg_ag2icmp_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2icmp_rule
for each row
execute function sgroups.check_rule_icmp_entries_no_overlap_trg();

drop trigger if exists trg_ag2icmp_rule_immutable_fields on sgroups.tbl_ag2icmp_rule;
create trigger trg_ag2icmp_rule_immutable_fields
before update on sgroups.tbl_ag2icmp_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2icmp_rule_resource_version on sgroups.tbl_ag2icmp_rule;
create trigger trg_ag2icmp_rule_resource_version
before insert or update on sgroups.tbl_ag2icmp_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2icmp_rule_outbox_au on sgroups.tbl_ag2icmp_rule;
create trigger trg_ag2icmp_rule_outbox_au
after insert or update on sgroups.tbl_ag2icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2IcmpRule', 'sgroups.vu_ag2icmp_rule');

drop trigger if exists trg_ag2icmp_rule_outbox_bd on sgroups.tbl_ag2icmp_rule;
create trigger trg_ag2icmp_rule_outbox_bd
before delete on sgroups.tbl_ag2icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2IcmpRule', 'sgroups.vu_ag2icmp_rule');


drop view if exists sgroups.vu_ag2icmp_rule cascade;
create or replace view sgroups.vu_ag2icmp_rule as
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

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2icmp_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns;


drop function if exists sgroups.list_ag2icmp_rule() cascade;
create or replace function sgroups.list_ag2icmp_rule()
returns setof sgroups.vu_ag2icmp_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2icmp_rule;
end;
$$ language plpgsql stable;


drop function if exists sgroups.resolve_local_ag_target(text, sgroups.endpoint) cascade;
create or replace function sgroups.resolve_local_ag_target(
  ruleNs text, agl sgroups.endpoint
)
returns bigint
as $$
declare
  aglID bigint;
  aglName text;
  aglNs text;
begin
    if agl is null then
      raise exception 'sgroups.resolve_local_ag_target: local address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing local address group reference';
    end if;


    aglName := nullif(btrim((agl).name), '');
    aglNs := nullif(btrim((agl).namespace), '');

    if aglName is null or aglNs is null then
      raise exception 'sgroups.resolve_local_ag_target: both local address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing local address group name and namespace';
    end if;

    if ruleNs <> aglNs then
      raise exception 'sgroups.resolve_local_ag_target: rule namespace must match local address group namespace'
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

    return aglID;
end;
$$ language plpgsql;



drop type if exists sgroups.row_of__ag2icmp_rule cascade;
create type sgroups.row_of__ag2icmp_rule as (
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
  ag_local sgroups.endpoint
);


drop function if exists sgroups.sync_ag2icmp_rule(sgroups.sync_op, sgroups.row_of__ag2icmp_rule) cascade;
create or replace function sgroups.sync_ag2icmp_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2icmp_rule
)
returns setof sgroups.vu_ag2icmp_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  aglID bigint;
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
      raise exception 'sgroups.sync_ag2icmp_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2icmp_rule: namespace is required for op=ups'
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
        aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
        insert into sgroups.tbl_ag2icmp_rule(
          uid, name, ns, agLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          aglID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'ag2icmp rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2icmp_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
      update sgroups.tbl_ag2icmp_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = aglID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'ag2icmp rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2icmp_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2icmp_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2icmp_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2icmp_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2icmp_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- AG2CIDR RULE -------------------------------------


drop table if exists sgroups.tbl_ag2cidr_rule cascade;
create table sgroups.tbl_ag2cidr_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    agLocal bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto,
    ip_v sgroups.ip_family not null,
    entries sgroups.port_entries[],
    cidr cidr not null,

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2cidr_rule_uid_uq  unique (uid),
    constraint ag2cidr_rule_name_uq unique (name, ns),
    constraint fk_ag2cidr_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2cidr_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2cidr_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists ag2cidr_rule_labels_gin_idx
  on sgroups.tbl_ag2cidr_rule using gin (labels);

create index if not exists ag2cidr_rule_annotations_gin_idx
  on sgroups.tbl_ag2cidr_rule using gin (annotations);

drop trigger if exists trg_ag2cidr_rule_entries_no_overlap on sgroups.tbl_ag2cidr_rule;
create trigger trg_ag2cidr_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2cidr_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_ag2cidr_rule_immutable_fields on sgroups.tbl_ag2cidr_rule;
create trigger trg_ag2cidr_rule_immutable_fields
before update on sgroups.tbl_ag2cidr_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2cidr_rule_resource_version on sgroups.tbl_ag2cidr_rule;
create trigger trg_ag2cidr_rule_resource_version
before insert or update on sgroups.tbl_ag2cidr_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2cidr_rule_outbox_au on sgroups.tbl_ag2cidr_rule;
create trigger trg_ag2cidr_rule_outbox_au
after insert or update on sgroups.tbl_ag2cidr_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2CidrRule', 'sgroups.vu_ag2cidr_rule');

drop trigger if exists trg_ag2cidr_rule_outbox_bd on sgroups.tbl_ag2cidr_rule;
create trigger trg_ag2cidr_rule_outbox_bd
before delete on sgroups.tbl_ag2cidr_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2CidrRule', 'sgroups.vu_ag2cidr_rule');


drop view if exists sgroups.vu_ag2cidr_rule cascade;
create or replace view sgroups.vu_ag2cidr_rule as
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
  r.cidr,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agl.name::text, ''),
    coalesce(agl_ns.name::text, ''),
    coalesce(agl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2cidr_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns;


drop function if exists sgroups.list_ag2cidr_rule() cascade;
create or replace function sgroups.list_ag2cidr_rule()
returns setof sgroups.vu_ag2cidr_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2cidr_rule;
end;
$$ language plpgsql stable;



drop type if exists sgroups.row_of__ag2cidr_rule cascade;
create type sgroups.row_of__ag2cidr_rule as (
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
  cidr cidr,
  ag_local sgroups.endpoint
);


drop function if exists sgroups.sync_ag2cidr_rule(sgroups.sync_op, sgroups.row_of__ag2cidr_rule) cascade;
create or replace function sgroups.sync_ag2cidr_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2cidr_rule
)
returns setof sgroups.vu_ag2cidr_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  aglID bigint;
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
      raise exception 'sgroups.sync_ag2cidr_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2cidr_rule: namespace is required for op=ups'
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
        aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
        insert into sgroups.tbl_ag2cidr_rule(
          uid, name, ns, agLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries, cidr, proto
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          aglID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries,
          (d).cidr,
          (d).proto
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'ag2cidr rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2cidr_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
      update sgroups.tbl_ag2cidr_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = aglID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries,
             cidr = (d).cidr,
             proto = (d).proto
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'ag2cidr rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2cidr_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2cidr_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2cidr_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2cidr_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2cidr_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- AG2CIDR ICMP RULE -------------------------------------


drop table if exists sgroups.tbl_ag2cidr_icmp_rule cascade;
create table sgroups.tbl_ag2cidr_icmp_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    agLocal bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    ip_v sgroups.ip_family not null,
    entries sgroups.icmp_entries[],
    cidr cidr not null,

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2cidr_icmp_rule_uid_uq  unique (uid),
    constraint ag2cidr_icmp_rule_name_uq unique (name, ns),
    constraint fk_ag2cidr_icmp_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2cidr_icmp_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2cidr_icmp_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists ag2cidr_icmp_rule_labels_gin_idx
  on sgroups.tbl_ag2cidr_icmp_rule using gin (labels);

create index if not exists ag2cidr_icmp_rule_annotations_gin_idx
  on sgroups.tbl_ag2cidr_icmp_rule using gin (annotations);

drop trigger if exists trg_ag2cidr_icmp_rule_entries_no_overlap on sgroups.tbl_ag2cidr_icmp_rule;
create trigger trg_ag2cidr_icmp_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2cidr_icmp_rule
for each row
execute function sgroups.check_rule_icmp_entries_no_overlap_trg();

drop trigger if exists trg_ag2cidr_icmp_rule_immutable_fields on sgroups.tbl_ag2cidr_icmp_rule;
create trigger trg_ag2cidr_icmp_rule_immutable_fields
before update on sgroups.tbl_ag2cidr_icmp_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2cidr_icmp_rule_resource_version on sgroups.tbl_ag2cidr_icmp_rule;
create trigger trg_ag2cidr_icmp_rule_resource_version
before insert or update on sgroups.tbl_ag2cidr_icmp_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2cidr_icmp_rule_outbox_au on sgroups.tbl_ag2cidr_icmp_rule;
create trigger trg_ag2cidr_icmp_rule_outbox_au
after insert or update on sgroups.tbl_ag2cidr_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2CidrIcmpRule', 'sgroups.vu_ag2cidr_icmp_rule');

drop trigger if exists trg_ag2cidr_icmp_rule_outbox_bd on sgroups.tbl_ag2cidr_icmp_rule;
create trigger trg_ag2cidr_icmp_rule_outbox_bd
before delete on sgroups.tbl_ag2cidr_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2CidrIcmpRule', 'sgroups.vu_ag2cidr_icmp_rule');


drop view if exists sgroups.vu_ag2cidr_icmp_rule cascade;
create or replace view sgroups.vu_ag2cidr_icmp_rule as
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
  r.cidr,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agl.name::text, ''),
    coalesce(agl_ns.name::text, ''),
    coalesce(agl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2cidr_icmp_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns;


drop function if exists sgroups.list_ag2cidr_icmp_rule() cascade;
create or replace function sgroups.list_ag2cidr_icmp_rule()
returns setof sgroups.vu_ag2cidr_icmp_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2cidr_icmp_rule;
end;
$$ language plpgsql stable;



drop type if exists sgroups.row_of__ag2cidr_icmp_rule cascade;
create type sgroups.row_of__ag2cidr_icmp_rule as (
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
  cidr cidr,
  ag_local sgroups.endpoint
);


drop function if exists sgroups.sync_ag2cidr_icmp_rule(sgroups.sync_op, sgroups.row_of__ag2cidr_icmp_rule) cascade;
create or replace function sgroups.sync_ag2cidr_icmp_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2cidr_icmp_rule
)
returns setof sgroups.vu_ag2cidr_icmp_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  aglID bigint;
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
      raise exception 'sgroups.sync_ag2cidr_icmp_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2cidr_icmp_rule: namespace is required for op=ups'
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
        aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
        insert into sgroups.tbl_ag2cidr_icmp_rule(
          uid, name, ns, agLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries, cidr
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          aglID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries,
          (d).cidr
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'ag2cidr icmp rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2cidr_icmp_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
      update sgroups.tbl_ag2cidr_icmp_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = aglID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries,
             cidr = (d).cidr
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'ag2cidr icmp rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2cidr_icmp_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2cidr_icmp_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2cidr_icmp_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2cidr_icmp_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2cidr_icmp_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

create extension if not exists citext;

---------------------------------- AG2FQDN RULE -------------------------------------

drop domain if exists sgroups.fqdn cascade;

create domain sgroups.fqdn
           as citext
   constraint fqdn_pattern
        check (
           value ~ '^([a-z0-9\*][a-z0-9_-]{1,62}){1}(\.[a-z0-9_][a-z0-9_-]{0,62})*$'
        )
   constraint fqdn_length
        check (
           length(value) < 256
        );


drop table if exists sgroups.tbl_ag2fqdn_rule cascade;
create table sgroups.tbl_ag2fqdn_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    agLocal bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto,
    ip_v sgroups.ip_family not null,
    entries sgroups.port_entries[],
    fqdn sgroups.fqdn not null,

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2fqdn_rule_uid_uq  unique (uid),
    constraint ag2fqdn_rule_name_uq unique (name, ns),
    constraint fk_ag2fqdn_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2fqdn_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2fqdn_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists ag2fqdn_rule_labels_gin_idx
  on sgroups.tbl_ag2fqdn_rule using gin (labels);

create index if not exists ag2fqdn_rule_annotations_gin_idx
  on sgroups.tbl_ag2fqdn_rule using gin (annotations);

drop trigger if exists trg_ag2fqdn_rule_entries_no_overlap on sgroups.tbl_ag2fqdn_rule;
create trigger trg_ag2fqdn_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2fqdn_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_ag2fqdn_rule_immutable_fields on sgroups.tbl_ag2fqdn_rule;
create trigger trg_ag2fqdn_rule_immutable_fields
before update on sgroups.tbl_ag2fqdn_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2fqdn_rule_resource_version on sgroups.tbl_ag2fqdn_rule;
create trigger trg_ag2fqdn_rule_resource_version
before insert or update on sgroups.tbl_ag2fqdn_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2fqdn_rule_outbox_au on sgroups.tbl_ag2fqdn_rule;
create trigger trg_ag2fqdn_rule_outbox_au
after insert or update on sgroups.tbl_ag2fqdn_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2FqdnRule', 'sgroups.vu_ag2fqdn_rule');

drop trigger if exists trg_ag2fqdn_rule_outbox_bd on sgroups.tbl_ag2fqdn_rule;
create trigger trg_ag2fqdn_rule_outbox_bd
before delete on sgroups.tbl_ag2fqdn_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2FqdnRule', 'sgroups.vu_ag2fqdn_rule');


drop view if exists sgroups.vu_ag2fqdn_rule cascade;
create or replace view sgroups.vu_ag2fqdn_rule as
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
  r.fqdn,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agl.name::text, ''),
    coalesce(agl_ns.name::text, ''),
    coalesce(agl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_ag2fqdn_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns;


drop function if exists sgroups.list_ag2fqdn_rule() cascade;
create or replace function sgroups.list_ag2fqdn_rule()
returns setof sgroups.vu_ag2fqdn_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2fqdn_rule;
end;
$$ language plpgsql stable;



drop type if exists sgroups.row_of__ag2fqdn_rule cascade;
create type sgroups.row_of__ag2fqdn_rule as (
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
  fqdn sgroups.fqdn,
  ag_local sgroups.endpoint
);


drop function if exists sgroups.sync_ag2fqdn_rule(sgroups.sync_op, sgroups.row_of__ag2fqdn_rule) cascade;
create or replace function sgroups.sync_ag2fqdn_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2fqdn_rule
)
returns setof sgroups.vu_ag2fqdn_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  aglID bigint;
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
      raise exception 'sgroups.sync_ag2fqdn_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2fqdn_rule: namespace is required for op=ups'
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
        aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
        insert into sgroups.tbl_ag2fqdn_rule(
          uid, name, ns, agLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries, fqdn, proto
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          aglID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries,
          (d).fqdn,
          (d).proto
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'ag2fqdn rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag2fqdn_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      aglID := sgroups.resolve_local_ag_target(norm_ns, (d).ag_local);
      update sgroups.tbl_ag2fqdn_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = aglID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries,
             fqdn = (d).fqdn,
             proto = (d).proto
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'ag2fqdn rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_ag2fqdn_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2fqdn_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2fqdn_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2fqdn_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2fqdn_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- SVC2SVC RULE -------------------------------------

drop table if exists sgroups.tbl_svc2svc_rule cascade;
create table sgroups.tbl_svc2svc_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    svcLocal bigint,
    svcRemote bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto,
    ip_v sgroups.ip_family,
    entries sgroups.port_entries[],

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint svc2svc_rule_uid_uq  unique (uid),
    constraint svc2svc_rule_name_uq unique (name, ns),
    constraint fk_svc2svc_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_svc2svc_rule___svc_local
       foreign key(svcLocal) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2svc_rule___svc_remote
       foreign key(svcRemote) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2svc_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists svc2svc_rule_labels_gin_idx
  on sgroups.tbl_svc2svc_rule using gin (labels);

create index if not exists svc2svc_rule_annotations_gin_idx
  on sgroups.tbl_svc2svc_rule using gin (annotations);

drop trigger if exists trg_svc2svc_rule_entries_no_overlap on sgroups.tbl_svc2svc_rule;
create trigger trg_svc2svc_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_svc2svc_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_svc2svc_rule_immutable_fields on sgroups.tbl_svc2svc_rule;
create trigger trg_svc2svc_rule_immutable_fields
before update on sgroups.tbl_svc2svc_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_svc2svc_rule_resource_version on sgroups.tbl_svc2svc_rule;
create trigger trg_svc2svc_rule_resource_version
before insert or update on sgroups.tbl_svc2svc_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_svc2svc_rule_outbox_au on sgroups.tbl_svc2svc_rule;
create trigger trg_svc2svc_rule_outbox_au
after insert or update on sgroups.tbl_svc2svc_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2SvcRule', 'sgroups.vu_svc2svc_rule');

drop trigger if exists trg_svc2svc_rule_outbox_bd on sgroups.tbl_svc2svc_rule;
create trigger trg_svc2svc_rule_outbox_bd
before delete on sgroups.tbl_svc2svc_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2SvcRule', 'sgroups.vu_svc2svc_rule');


drop view if exists sgroups.vu_svc2svc_rule cascade;
create or replace view sgroups.vu_svc2svc_rule as
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
  coalesce(r.ip_v::text, '')  as ip_v,
  r.entries,
  coalesce(r.proto::text, '') as proto,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcl.name::text, ''),
    coalesce(svcl_ns.name::text, ''),
    coalesce(svcl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcr.name::text, ''),
    coalesce(svcr_ns.name::text, ''),
    coalesce(svcr.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as remote,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_svc2svc_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_service svcl on svcl.id = r.svcLocal
left join sgroups.tbl_namespace svcl_ns on svcl_ns.id = svcl.ns
left join sgroups.tbl_service svcr on svcr.id = r.svcRemote
left join sgroups.tbl_namespace svcr_ns on svcr_ns.id = svcr.ns;


drop function if exists sgroups.list_svc2svc_rule() cascade;
create or replace function sgroups.list_svc2svc_rule()
returns setof sgroups.vu_svc2svc_rule
as $$
begin
  return query
  select * from sgroups.vu_svc2svc_rule;
end;
$$ language plpgsql stable;


drop function if exists sgroups.resolve_svc2svc_rule_targets(text, sgroups.endpoint, sgroups.endpoint) cascade;
create or replace function sgroups.resolve_svc2svc_rule_targets(
  ruleNs text, svcl sgroups.endpoint, svcr sgroups.endpoint
)
returns table(svcl_id bigint, svcr_id bigint)
as $$
declare
  svclID bigint;
  svclName text;
  svclNs text;
  svcrID bigint;
  svcrName text;
  svcrNs text;
begin
    if svcl is null then
      raise exception 'sgroups.resolve_svc2svc_rule_targets: local service reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing local service reference';
    end if;
    if svcr is null then
      raise exception 'sgroups.resolve_svc2svc_rule_targets: remote service reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing remote service reference';
    end if;

    svclName := nullif(btrim((svcl).name), '');
    svclNs := nullif(btrim((svcl).namespace), '');
    svcrName := nullif(btrim((svcr).name), '');
    svcrNs := nullif(btrim((svcr).namespace), '');

    if svclName is null or svclNs is null then
      raise exception 'sgroups.resolve_svc2svc_rule_targets: both local service name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing local service name and namespace';
    end if;
    if svcrName is null or svcrNs is null then
      raise exception 'sgroups.resolve_svc2svc_rule_targets: both remote service name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing remote service name and namespace';
    end if;

    if ruleNs <> svclNs then
      raise exception 'sgroups.resolve_svc2svc_rule_targets: rule namespace must match local service namespace'
        using detail = 'SG0007', hint = 'pass rule namespace that matches local service namespace';
    end if;

    select t.id
    from sgroups.tbl_service t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = svclName::sgroups.rname
      and ns.name = svclNs::sgroups.rname
    into svclID;
    if svclID is null then
      raise exception 'local service not found for name=% namespace=%', svclName, svclNs
        using detail = 'SG0009', hint = 'pass existing local service name and namespace';
    end if;

    select t.id
    from sgroups.tbl_service t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = svcrName::sgroups.rname
      and ns.name = svcrNs::sgroups.rname
    into svcrID;
    if svcrID is null then
      raise exception 'remote service not found for name=% namespace=%', svcrName, svcrNs
        using detail = 'SG0009', hint = 'pass existing remote service name and namespace';
    end if;

    svcl_id := svclID;
    svcr_id := svcrID;
    return next;
end;
$$ language plpgsql;


drop type if exists sgroups.row_of__svc2svc_rule cascade;
create type sgroups.row_of__svc2svc_rule as (
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
  svc_local sgroups.endpoint,
  svc_remote sgroups.endpoint
);


drop function if exists sgroups.sync_svc2svc_rule(sgroups.sync_op, sgroups.row_of__svc2svc_rule) cascade;
create or replace function sgroups.sync_svc2svc_rule(
  op sgroups.sync_op, d sgroups.row_of__svc2svc_rule
)
returns setof sgroups.vu_svc2svc_rule
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
      raise exception 'sgroups.sync_svc2svc_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_svc2svc_rule: namespace is required for op=ups'
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
          select * from sgroups.resolve_svc2svc_rule_targets(norm_ns, (d).svc_local, (d).svc_remote)
        )
        insert into sgroups.tbl_svc2svc_rule(
          uid, name, ns, svcLocal, svcRemote, display_name, labels, annotations, comment, description, action, traffic, proto, ip_v, entries
        )
        select
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          ids.svcl_id,
          ids.svcr_id,
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
      from sgroups.vu_svc2svc_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_svc2svc_rule_targets(norm_ns, (d).svc_local, (d).svc_remote)
      )
      update sgroups.tbl_svc2svc_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             svcLocal = ids.svcl_id,
             svcRemote = ids.svcr_id,
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
      if exists (select 1 from sgroups.tbl_svc2svc_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_svc2svc_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_svc2svc_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_svc2svc_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_svc2svc_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- SVC2FQDN RULE -------------------------------------


drop table if exists sgroups.tbl_svc2fqdn_rule cascade;
create table sgroups.tbl_svc2fqdn_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    svcLocal bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto,
    ip_v sgroups.ip_family not null,
    entries sgroups.port_entries[],
    fqdn sgroups.fqdn not null,

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint svc2fqdn_rule_uid_uq  unique (uid),
    constraint svc2fqdn_rule_name_uq unique (name, ns),
    constraint fk_svc2fqdn_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_svc2fqdn_rule___svc_local
       foreign key(svcLocal) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2fqdn_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists svc2fqdn_rule_labels_gin_idx
  on sgroups.tbl_svc2fqdn_rule using gin (labels);

create index if not exists svc2fqdn_rule_annotations_gin_idx
  on sgroups.tbl_svc2fqdn_rule using gin (annotations);

drop trigger if exists trg_svc2fqdn_rule_entries_no_overlap on sgroups.tbl_svc2fqdn_rule;
create trigger trg_svc2fqdn_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_svc2fqdn_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_svc2fqdn_rule_immutable_fields on sgroups.tbl_svc2fqdn_rule;
create trigger trg_svc2fqdn_rule_immutable_fields
before update on sgroups.tbl_svc2fqdn_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_svc2fqdn_rule_resource_version on sgroups.tbl_svc2fqdn_rule;
create trigger trg_svc2fqdn_rule_resource_version
before insert or update on sgroups.tbl_svc2fqdn_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_svc2fqdn_rule_outbox_au on sgroups.tbl_svc2fqdn_rule;
create trigger trg_svc2fqdn_rule_outbox_au
after insert or update on sgroups.tbl_svc2fqdn_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2FqdnRule', 'sgroups.vu_svc2fqdn_rule');

drop trigger if exists trg_svc2fqdn_rule_outbox_bd on sgroups.tbl_svc2fqdn_rule;
create trigger trg_svc2fqdn_rule_outbox_bd
before delete on sgroups.tbl_svc2fqdn_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2FqdnRule', 'sgroups.vu_svc2fqdn_rule');


drop view if exists sgroups.vu_svc2fqdn_rule cascade;
create or replace view sgroups.vu_svc2fqdn_rule as
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
  r.fqdn,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcl.name::text, ''),
    coalesce(svcl_ns.name::text, ''),
    coalesce(svcl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_svc2fqdn_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_service svcl on svcl.id = r.svcLocal
left join sgroups.tbl_namespace svcl_ns on svcl_ns.id = svcl.ns;


drop function if exists sgroups.list_svc2fqdn_rule() cascade;
create or replace function sgroups.list_svc2fqdn_rule()
returns setof sgroups.vu_svc2fqdn_rule
as $$
begin
  return query
  select * from sgroups.vu_svc2fqdn_rule;
end;
$$ language plpgsql stable;


drop function if exists sgroups.resolve_local_svc_target(text, sgroups.endpoint) cascade;
create or replace function sgroups.resolve_local_svc_target(
  ruleNs text, svcl sgroups.endpoint
)
returns bigint
as $$
declare
  svclID bigint;
  svclName text;
  svclNs text;
begin
    if svcl is null then
      raise exception 'sgroups.resolve_local_svc_target: local service reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing local service reference';
    end if;


    svclName := nullif(btrim((svcl).name), '');
    svclNs := nullif(btrim((svcl).namespace), '');

    if svclName is null or svclNs is null then
      raise exception 'sgroups.resolve_local_svc_target: both local service name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing local service name and namespace';
    end if;

    if ruleNs <> svclNs then
      raise exception 'sgroups.resolve_local_svc_target: rule namespace must match local service namespace'
        using detail = 'SG0007', hint = 'pass rule namespace that matches local service namespace';
    end if;

    select t.id
    from sgroups.tbl_service t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = svclName::sgroups.rname
      and ns.name = svclNs::sgroups.rname
    into svclID;
    if svclID is null then
      raise exception 'local service not found for name=% namespace=%', svclName, svclNs
        using detail = 'SG0009', hint = 'pass existing local service name and namespace';
    end if;

    return svclID;
end;
$$ language plpgsql;


drop type if exists sgroups.row_of__svc2fqdn_rule cascade;
create type sgroups.row_of__svc2fqdn_rule as (
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
  fqdn sgroups.fqdn,
  svc_local sgroups.endpoint
);


drop function if exists sgroups.sync_svc2fqdn_rule(sgroups.sync_op, sgroups.row_of__svc2fqdn_rule) cascade;
create or replace function sgroups.sync_svc2fqdn_rule(
  op sgroups.sync_op, d sgroups.row_of__svc2fqdn_rule
)
returns setof sgroups.vu_svc2fqdn_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  svclID bigint;
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
      raise exception 'sgroups.sync_svc2fqdn_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_svc2fqdn_rule: namespace is required for op=ups'
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
        svclID := sgroups.resolve_local_svc_target(norm_ns, (d).svc_local);
        insert into sgroups.tbl_svc2fqdn_rule(
          uid, name, ns, svcLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries, fqdn, proto
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          svclID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries,
          (d).fqdn,
          (d).proto
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'svc2fqdn rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_svc2fqdn_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      svclID := sgroups.resolve_local_svc_target(norm_ns, (d).svc_local);
      update sgroups.tbl_svc2fqdn_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             svcLocal = svclID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries,
             fqdn = (d).fqdn,
             proto = (d).proto
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'svc2fqdn rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_svc2fqdn_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_svc2fqdn_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_svc2fqdn_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_svc2fqdn_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_svc2fqdn_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- SVC2CIDR RULE -------------------------------------


drop table if exists sgroups.tbl_svc2cidr_rule cascade;
create table sgroups.tbl_svc2cidr_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    svcLocal bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto,
    ip_v sgroups.ip_family,
    entries sgroups.port_entries[],
    cidr cidr not null,

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint svc2cidr_rule_uid_uq  unique (uid),
    constraint svc2cidr_rule_name_uq unique (name, ns),
    constraint fk_svc2cidr_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_svc2cidr_rule___svc_local
       foreign key(svcLocal) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2cidr_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists svc2cidr_rule_labels_gin_idx
  on sgroups.tbl_svc2cidr_rule using gin (labels);

create index if not exists svc2cidr_rule_annotations_gin_idx
  on sgroups.tbl_svc2cidr_rule using gin (annotations);

drop trigger if exists trg_svc2cidr_rule_entries_no_overlap on sgroups.tbl_svc2cidr_rule;
create trigger trg_svc2cidr_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_svc2cidr_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_svc2cidr_rule_immutable_fields on sgroups.tbl_svc2cidr_rule;
create trigger trg_svc2cidr_rule_immutable_fields
before update on sgroups.tbl_svc2cidr_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_svc2cidr_rule_resource_version on sgroups.tbl_svc2cidr_rule;
create trigger trg_svc2cidr_rule_resource_version
before insert or update on sgroups.tbl_svc2cidr_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_svc2cidr_rule_outbox_au on sgroups.tbl_svc2cidr_rule;
create trigger trg_svc2cidr_rule_outbox_au
after insert or update on sgroups.tbl_svc2cidr_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2CidrRule', 'sgroups.vu_svc2cidr_rule');

drop trigger if exists trg_svc2cidr_rule_outbox_bd on sgroups.tbl_svc2cidr_rule;
create trigger trg_svc2cidr_rule_outbox_bd
before delete on sgroups.tbl_svc2cidr_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2CidrRule', 'sgroups.vu_svc2cidr_rule');


drop view if exists sgroups.vu_svc2cidr_rule cascade;
create or replace view sgroups.vu_svc2cidr_rule as
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
  coalesce(r.ip_v::text, '')  as ip_v,
  r.entries,
  coalesce(r.proto::text, '') as proto,
  r.cidr,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcl.name::text, ''),
    coalesce(svcl_ns.name::text, ''),
    coalesce(svcl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_svc2cidr_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_service svcl on svcl.id = r.svcLocal
left join sgroups.tbl_namespace svcl_ns on svcl_ns.id = svcl.ns;


drop function if exists sgroups.list_svc2cidr_rule() cascade;
create or replace function sgroups.list_svc2cidr_rule()
returns setof sgroups.vu_svc2cidr_rule
as $$
begin
  return query
  select * from sgroups.vu_svc2cidr_rule;
end;
$$ language plpgsql stable;



drop type if exists sgroups.row_of__svc2cidr_rule cascade;
create type sgroups.row_of__svc2cidr_rule as (
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
  cidr cidr,
  svc_local sgroups.endpoint
);


drop function if exists sgroups.sync_svc2cidr_rule(sgroups.sync_op, sgroups.row_of__svc2cidr_rule) cascade;
create or replace function sgroups.sync_svc2cidr_rule(
  op sgroups.sync_op, d sgroups.row_of__svc2cidr_rule
)
returns setof sgroups.vu_svc2cidr_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  svclID bigint;
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
      raise exception 'sgroups.sync_svc2cidr_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_svc2cidr_rule: namespace is required for op=ups'
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
        svclID := sgroups.resolve_local_svc_target(norm_ns, (d).svc_local);
        insert into sgroups.tbl_svc2cidr_rule(
          uid, name, ns, svcLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries, cidr, proto
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          svclID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries,
          (d).cidr,
          (d).proto
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'svc2cidr rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_svc2cidr_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      svclID := sgroups.resolve_local_svc_target(norm_ns, (d).svc_local);
      update sgroups.tbl_svc2cidr_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             svcLocal = svclID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries,
             cidr = (d).cidr,
             proto = (d).proto
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'svc2cidr rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_svc2cidr_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_svc2cidr_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_svc2cidr_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_svc2cidr_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_svc2cidr_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

---------------------------------- SVC2CIDR ICMP RULE -------------------------------------


drop table if exists sgroups.tbl_svc2cidr_icmp_rule cascade;
create table sgroups.tbl_svc2cidr_icmp_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    svcLocal bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    ip_v sgroups.ip_family not null,
    entries sgroups.icmp_entries[],
    cidr cidr not null,

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint svc2cidr_icmp_rule_uid_uq  unique (uid),
    constraint svc2cidr_icmp_rule_name_uq unique (name, ns),
    constraint fk_svc2cidr_icmp_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_svc2cidr_icmp_rule___svc_local
       foreign key(svcLocal) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2cidr_icmp_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);


create index if not exists svc2cidr_icmp_rule_labels_gin_idx
  on sgroups.tbl_svc2cidr_icmp_rule using gin (labels);

create index if not exists svc2cidr_icmp_rule_annotations_gin_idx
  on sgroups.tbl_svc2cidr_icmp_rule using gin (annotations);

drop trigger if exists trg_svc2cidr_icmp_rule_entries_no_overlap on sgroups.tbl_svc2cidr_icmp_rule;
create trigger trg_svc2cidr_icmp_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_svc2cidr_icmp_rule
for each row
execute function sgroups.check_rule_icmp_entries_no_overlap_trg();

drop trigger if exists trg_svc2cidr_icmp_rule_immutable_fields on sgroups.tbl_svc2cidr_icmp_rule;
create trigger trg_svc2cidr_icmp_rule_immutable_fields
before update on sgroups.tbl_svc2cidr_icmp_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_svc2cidr_icmp_rule_resource_version on sgroups.tbl_svc2cidr_icmp_rule;
create trigger trg_svc2cidr_icmp_rule_resource_version
before insert or update on sgroups.tbl_svc2cidr_icmp_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_svc2cidr_icmp_rule_outbox_au on sgroups.tbl_svc2cidr_icmp_rule;
create trigger trg_svc2cidr_icmp_rule_outbox_au
after insert or update on sgroups.tbl_svc2cidr_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2CidrIcmpRule', 'sgroups.vu_svc2cidr_icmp_rule');

drop trigger if exists trg_svc2cidr_icmp_rule_outbox_bd on sgroups.tbl_svc2cidr_icmp_rule;
create trigger trg_svc2cidr_icmp_rule_outbox_bd
before delete on sgroups.tbl_svc2cidr_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2CidrIcmpRule', 'sgroups.vu_svc2cidr_icmp_rule');


drop view if exists sgroups.vu_svc2cidr_icmp_rule cascade;
create or replace view sgroups.vu_svc2cidr_icmp_rule as
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
  r.cidr,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcl.name::text, ''),
    coalesce(svcl_ns.name::text, ''),
    coalesce(svcl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_svc2cidr_icmp_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_service svcl on svcl.id = r.svcLocal
left join sgroups.tbl_namespace svcl_ns on svcl_ns.id = svcl.ns;


drop function if exists sgroups.list_svc2cidr_icmp_rule() cascade;
create or replace function sgroups.list_svc2cidr_icmp_rule()
returns setof sgroups.vu_svc2cidr_icmp_rule
as $$
begin
  return query
  select * from sgroups.vu_svc2cidr_icmp_rule;
end;
$$ language plpgsql stable;



drop type if exists sgroups.row_of__svc2cidr_icmp_rule cascade;
create type sgroups.row_of__svc2cidr_icmp_rule as (
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
  cidr cidr,
  svc_local sgroups.endpoint
);


drop function if exists sgroups.sync_svc2cidr_icmp_rule(sgroups.sync_op, sgroups.row_of__svc2cidr_icmp_rule) cascade;
create or replace function sgroups.sync_svc2cidr_icmp_rule(
  op sgroups.sync_op, d sgroups.row_of__svc2cidr_icmp_rule
)
returns setof sgroups.vu_svc2cidr_icmp_rule
as $$
declare
  affected_uid uuid;
  rl_uid uuid;
  ruleID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
  svclID bigint;
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
      raise exception 'sgroups.sync_svc2cidr_icmp_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_svc2cidr_icmp_rule: namespace is required for op=ups'
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
        svclID := sgroups.resolve_local_svc_target(norm_ns, (d).svc_local);
        insert into sgroups.tbl_svc2cidr_icmp_rule(
          uid, name, ns, svcLocal, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries, cidr
        )
        values (
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          svclID,
          (d).display_name,
          (d).labels,
          (d).annotations,
          (d).comment,
          (d).description,
          (d).action,
          (d).traffic,
          (d).ip_v,
          (d).entries,
          (d).cidr
        )
        returning id, uid into ruleID, affected_uid;
      exception
        when unique_violation then
          raise exception 'svc2cidr icmp rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_svc2cidr_icmp_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      svclID := sgroups.resolve_local_svc_target(norm_ns, (d).svc_local);
      update sgroups.tbl_svc2cidr_icmp_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             svcLocal = svclID,
             action = (d).action,
             traffic = (d).traffic,
             ip_v = (d).ip_v,
             entries = (d).entries,
             cidr = (d).cidr
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into ruleID;
    exception
      when unique_violation then
        raise exception 'svc2cidr icmp rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_svc2cidr_icmp_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_svc2cidr_icmp_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_svc2cidr_icmp_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_svc2cidr_icmp_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_svc2cidr_icmp_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

drop table if exists sgroups.tbl_ag2svc_rule cascade;
create table sgroups.tbl_ag2svc_rule (
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
    proto sgroups.proto,
    ip_v sgroups.ip_family,
    entries sgroups.port_entries[],

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ag2svc_rule_uid_uq  unique (uid),
    constraint ag2svc_rule_name_uq unique (name, ns),
    constraint fk_ag2svc_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_ag2svc_rule___ag_local
       foreign key(agLocal) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2svc_rule___svc_remote
       foreign key(svcRemote) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_ag2svc_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists ag2svc_rule_labels_gin_idx
  on sgroups.tbl_ag2svc_rule using gin (labels);

create index if not exists ag2svc_rule_annotations_gin_idx
  on sgroups.tbl_ag2svc_rule using gin (annotations);

drop trigger if exists trg_ag2svc_rule_entries_no_overlap on sgroups.tbl_ag2svc_rule;
create trigger trg_ag2svc_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_ag2svc_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_ag2svc_rule_immutable_fields on sgroups.tbl_ag2svc_rule;
create trigger trg_ag2svc_rule_immutable_fields
before update on sgroups.tbl_ag2svc_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag2svc_rule_resource_version on sgroups.tbl_ag2svc_rule;
create trigger trg_ag2svc_rule_resource_version
before insert or update on sgroups.tbl_ag2svc_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag2svc_rule_outbox_au on sgroups.tbl_ag2svc_rule;
create trigger trg_ag2svc_rule_outbox_au
after insert or update on sgroups.tbl_ag2svc_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2SvcRule', 'sgroups.vu_ag2svc_rule');

drop trigger if exists trg_ag2svc_rule_outbox_bd on sgroups.tbl_ag2svc_rule;
create trigger trg_ag2svc_rule_outbox_bd
before delete on sgroups.tbl_ag2svc_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Ag2SvcRule', 'sgroups.vu_ag2svc_rule');


drop view if exists sgroups.vu_ag2svc_rule cascade;
create or replace view sgroups.vu_ag2svc_rule as
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
  coalesce(r.ip_v::text, '')  as ip_v,
  coalesce(r.proto::text, '') as proto,
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
from sgroups.tbl_ag2svc_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_ag agl on agl.id = r.agLocal
left join sgroups.tbl_namespace agl_ns on agl_ns.id = agl.ns
left join sgroups.tbl_service svcr on svcr.id = r.svcRemote
left join sgroups.tbl_namespace svcr_ns on svcr_ns.id = svcr.ns;


drop function if exists sgroups.list_ag2svc_rule() cascade;
create or replace function sgroups.list_ag2svc_rule()
returns setof sgroups.vu_ag2svc_rule
as $$
begin
  return query
  select * from sgroups.vu_ag2svc_rule;
end;
$$ language plpgsql stable;


drop function if exists sgroups.resolve_ag2svc_rule_targets(text, sgroups.endpoint, sgroups.endpoint) cascade;
create or replace function sgroups.resolve_ag2svc_rule_targets(
  ruleNs text, agl sgroups.endpoint, svcr sgroups.endpoint
)
returns table(agl_id bigint, svcr_id bigint)
as $$
declare
  aglID bigint;
  aglName text;
  aglNs text;
  svcrID bigint;
  svcrName text;
  svcrNs text;
begin
    if agl is null then
      raise exception 'sgroups.resolve_ag2svc_rule_targets: local address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing local address group reference';
    end if;
    if svcr is null then
      raise exception 'sgroups.resolve_ag2svc_rule_targets: remote service reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing remote service reference';
    end if;

    aglName := nullif(btrim((agl).name), '');
    aglNs := nullif(btrim((agl).namespace), '');
    svcrName := nullif(btrim((svcr).name), '');
    svcrNs := nullif(btrim((svcr).namespace), '');

    if aglName is null or aglNs is null then
      raise exception 'sgroups.resolve_ag2svc_rule_targets: both local address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing local address group name and namespace';
    end if;
    if svcrName is null or svcrNs is null then
      raise exception 'sgroups.resolve_ag2svc_rule_targets: both remote service name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing remote service name and namespace';
    end if;

    if ruleNs <> aglNs then
      raise exception 'sgroups.resolve_ag2svc_rule_targets: rule namespace must match local address group namespace'
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
    from sgroups.tbl_service t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = svcrName::sgroups.rname
      and ns.name = svcrNs::sgroups.rname
    into svcrID;
    if svcrID is null then
      raise exception 'remote service not found for name=% namespace=%', svcrName, svcrNs
        using detail = 'SG0009', hint = 'pass existing remote service name and namespace';
    end if;

    agl_id := aglID;
    svcr_id := svcrID;
    return next;
end;
$$ language plpgsql;


drop type if exists sgroups.row_of__ag2svc_rule cascade;
create type sgroups.row_of__ag2svc_rule as (
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
  svc_remote sgroups.endpoint
);


drop function if exists sgroups.sync_ag2svc_rule(sgroups.sync_op, sgroups.row_of__ag2svc_rule) cascade;
create or replace function sgroups.sync_ag2svc_rule(
  op sgroups.sync_op, d sgroups.row_of__ag2svc_rule
)
returns setof sgroups.vu_ag2svc_rule
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
      raise exception 'sgroups.sync_ag2svc_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_ag2svc_rule: namespace is required for op=ups'
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
        insert into sgroups.tbl_ag2svc_rule(
          uid, name, ns, agLocal, svcRemote, display_name, labels, annotations,
          comment, description, action, traffic, proto, ip_v, entries
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
      from sgroups.vu_ag2svc_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_ag2svc_rule_targets(norm_ns, (d).ag_local, (d).svc_remote)
      )
      update sgroups.tbl_ag2svc_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             agLocal = ids.agl_id,
             svcRemote = ids.svcr_id,
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
      if exists (select 1 from sgroups.tbl_ag2svc_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_ag2svc_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_ag2svc_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag2svc_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_ag2svc_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

drop table if exists sgroups.tbl_svc2ag_rule cascade;
create table sgroups.tbl_svc2ag_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    svcLocal bigint,
    agRemote bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    action sgroups.policy_action not null default 'DENY',
    traffic sgroups.traffic not null,
    proto sgroups.proto,
    ip_v sgroups.ip_family,
    entries sgroups.port_entries[],

    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint svc2ag_rule_uid_uq  unique (uid),
    constraint svc2ag_rule_name_uq unique (name, ns),
    constraint fk_svc2ag_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_svc2ag_rule___svc_local
       foreign key(svcLocal) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2ag_rule___ag_remote
       foreign key(agRemote) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2ag_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists svc2ag_rule_labels_gin_idx
  on sgroups.tbl_svc2ag_rule using gin (labels);

create index if not exists svc2ag_rule_annotations_gin_idx
  on sgroups.tbl_svc2ag_rule using gin (annotations);

drop trigger if exists trg_svc2ag_rule_entries_no_overlap on sgroups.tbl_svc2ag_rule;
create trigger trg_svc2ag_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_svc2ag_rule
for each row
execute function sgroups.check_rule_port_entries_no_overlap_trg();

drop trigger if exists trg_svc2ag_rule_immutable_fields on sgroups.tbl_svc2ag_rule;
create trigger trg_svc2ag_rule_immutable_fields
before update on sgroups.tbl_svc2ag_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_svc2ag_rule_resource_version on sgroups.tbl_svc2ag_rule;
create trigger trg_svc2ag_rule_resource_version
before insert or update on sgroups.tbl_svc2ag_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_svc2ag_rule_outbox_au on sgroups.tbl_svc2ag_rule;
create trigger trg_svc2ag_rule_outbox_au
after insert or update on sgroups.tbl_svc2ag_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2AgRule', 'sgroups.vu_svc2ag_rule');

drop trigger if exists trg_svc2ag_rule_outbox_bd on sgroups.tbl_svc2ag_rule;
create trigger trg_svc2ag_rule_outbox_bd
before delete on sgroups.tbl_svc2ag_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2AgRule', 'sgroups.vu_svc2ag_rule');


drop view if exists sgroups.vu_svc2ag_rule cascade;
create or replace view sgroups.vu_svc2ag_rule as
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
  coalesce(r.ip_v::text, '')  as ip_v,
  r.entries,
  coalesce(r.proto::text, '') as proto,
  row(
    'Service'::sgroups.resource_type,
    coalesce(svcl.name::text, ''),
    coalesce(svcl_ns.name::text, ''),
    coalesce(svcl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agr.name::text, ''),
    coalesce(agr_ns.name::text, ''),
    coalesce(agr.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as remote,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_svc2ag_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_service svcl on svcl.id = r.svcLocal
left join sgroups.tbl_namespace svcl_ns on svcl_ns.id = svcl.ns
left join sgroups.tbl_ag agr on agr.id = r.agRemote
left join sgroups.tbl_namespace agr_ns on agr_ns.id = agr.ns;


drop function if exists sgroups.list_svc2ag_rule() cascade;
create or replace function sgroups.list_svc2ag_rule()
returns setof sgroups.vu_svc2ag_rule
as $$
begin
  return query
  select * from sgroups.vu_svc2ag_rule;
end;
$$ language plpgsql stable;


drop function if exists sgroups.resolve_svc2ag_rule_targets(text, sgroups.endpoint, sgroups.endpoint) cascade;
create or replace function sgroups.resolve_svc2ag_rule_targets(
  ruleNs text, svcl sgroups.endpoint, agr sgroups.endpoint
)
returns table(svcl_id bigint, agr_id bigint)
as $$
declare
  svclID bigint;
  svclName text;
  svclNs text;
  agrID bigint;
  agrName text;
  agrNs text;
begin
    if svcl is null then
      raise exception 'sgroups.resolve_svc2ag_rule_targets: local service reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing local service reference';
    end if;
    if agr is null then
      raise exception 'sgroups.resolve_svc2ag_rule_targets: remote address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing remote address group reference';
    end if;

    svclName := nullif(btrim((svcl).name), '');
    svclNs := nullif(btrim((svcl).namespace), '');
    agrName := nullif(btrim((agr).name), '');
    agrNs := nullif(btrim((agr).namespace), '');

    if svclName is null or svclNs is null then
      raise exception 'sgroups.resolve_svc2ag_rule_targets: both local service name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing local service name and namespace';
    end if;
    if agrName is null or agrNs is null then
      raise exception 'sgroups.resolve_svc2ag_rule_targets: both remote address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing remote address group name and namespace';
    end if;

    if ruleNs <> svclNs then
      raise exception 'sgroups.resolve_svc2ag_rule_targets: rule namespace must match local service namespace'
        using detail = 'SG0007', hint = 'pass rule namespace that matches local service namespace';
    end if;

    select t.id
    from sgroups.tbl_service t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = svclName::sgroups.rname
      and ns.name = svclNs::sgroups.rname
    into svclID;
    if svclID is null then
      raise exception 'local service not found for name=% namespace=%', svclName, svclNs
        using detail = 'SG0009', hint = 'pass existing local service name and namespace';
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

    svcl_id := svclID;
    agr_id := agrID;
    return next;
end;
$$ language plpgsql;


drop type if exists sgroups.row_of__svc2ag_rule cascade;
create type sgroups.row_of__svc2ag_rule as (
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
  svc_local sgroups.endpoint,
  ag_remote sgroups.endpoint
);


drop function if exists sgroups.sync_svc2ag_rule(sgroups.sync_op, sgroups.row_of__svc2ag_rule) cascade;
create or replace function sgroups.sync_svc2ag_rule(
  op sgroups.sync_op, d sgroups.row_of__svc2ag_rule
)
returns setof sgroups.vu_svc2ag_rule
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
      raise exception 'sgroups.sync_svc2ag_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_svc2ag_rule: namespace is required for op=ups'
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
          select * from sgroups.resolve_svc2ag_rule_targets(norm_ns, (d).svc_local, (d).ag_remote)
        )
        insert into sgroups.tbl_svc2ag_rule(
          uid, name, ns, svcLocal, agRemote, display_name, labels, annotations, comment, description, action, traffic, proto, ip_v, entries
        )
        select
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          ids.svcl_id,
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
      from sgroups.vu_svc2ag_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_svc2ag_rule_targets(norm_ns, (d).svc_local, (d).ag_remote)
      )
      update sgroups.tbl_svc2ag_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             svcLocal = ids.svcl_id,
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
      if exists (select 1 from sgroups.tbl_svc2ag_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_svc2ag_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_svc2ag_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_svc2ag_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_svc2ag_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

drop table if exists sgroups.tbl_svc2ag_icmp_rule cascade;
create table sgroups.tbl_svc2ag_icmp_rule (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    svcLocal bigint,
    agRemote bigint,
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

    constraint svc2ag_icmp_rule_uid_uq  unique (uid),
    constraint svc2ag_icmp_rule_name_uq unique (name, ns),
    constraint fk_svc2ag_icmp_rule___registry
      foreign key(uid) references sgroups.tbl_rule_registry(uid)
            on delete cascade
            deferrable initially deferred,
    constraint fk_svc2ag_icmp_rule___svc_local
       foreign key(svcLocal) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2ag_icmp_rule___ag_remote
       foreign key(agRemote) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_svc2ag_icmp_rule___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists svc2ag_icmp_rule_labels_gin_idx
  on sgroups.tbl_svc2ag_icmp_rule using gin (labels);

create index if not exists svc2ag_icmp_rule_annotations_gin_idx
  on sgroups.tbl_svc2ag_icmp_rule using gin (annotations);

drop trigger if exists trg_svc2ag_icmp_rule_entries_no_overlap on sgroups.tbl_svc2ag_icmp_rule;
create trigger trg_svc2ag_icmp_rule_entries_no_overlap
before insert or update of entries on sgroups.tbl_svc2ag_icmp_rule
for each row
execute function sgroups.check_rule_icmp_entries_no_overlap_trg();

drop trigger if exists trg_svc2ag_icmp_rule_immutable_fields on sgroups.tbl_svc2ag_icmp_rule;
create trigger trg_svc2ag_icmp_rule_immutable_fields
before update on sgroups.tbl_svc2ag_icmp_rule
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_svc2ag_icmp_rule_resource_version on sgroups.tbl_svc2ag_icmp_rule;
create trigger trg_svc2ag_icmp_rule_resource_version
before insert or update on sgroups.tbl_svc2ag_icmp_rule
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_svc2ag_icmp_rule_outbox_au on sgroups.tbl_svc2ag_icmp_rule;
create trigger trg_svc2ag_icmp_rule_outbox_au
after insert or update on sgroups.tbl_svc2ag_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2AgIcmpRule', 'sgroups.vu_svc2ag_icmp_rule');

drop trigger if exists trg_svc2ag_icmp_rule_outbox_bd on sgroups.tbl_svc2ag_icmp_rule;
create trigger trg_svc2ag_icmp_rule_outbox_bd
before delete on sgroups.tbl_svc2ag_icmp_rule
for each row
execute function sgroups.capture_outbox_resource_events_trg('Svc2AgIcmpRule', 'sgroups.vu_svc2ag_icmp_rule');


drop view if exists sgroups.vu_svc2ag_icmp_rule cascade;
create or replace view sgroups.vu_svc2ag_icmp_rule as
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
    'Service'::sgroups.resource_type,
    coalesce(svcl.name::text, ''),
    coalesce(svcl_ns.name::text, ''),
    coalesce(svcl.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as local,
  row(
    'AddressGroup'::sgroups.resource_type,
    coalesce(agr.name::text, ''),
    coalesce(agr_ns.name::text, ''),
    coalesce(agr.labels, hstore(array[]::text[], array[]::text[]))
  )::sgroups.endpoint as remote,

  r.creation_timestamp,
  r.resource_version
from sgroups.tbl_svc2ag_icmp_rule as r
left join sgroups.tbl_namespace ns on ns.id = r.ns
left join sgroups.tbl_service svcl on svcl.id = r.svcLocal
left join sgroups.tbl_namespace svcl_ns on svcl_ns.id = svcl.ns
left join sgroups.tbl_ag agr on agr.id = r.agRemote
left join sgroups.tbl_namespace agr_ns on agr_ns.id = agr.ns;


drop function if exists sgroups.list_svc2ag_icmp_rule() cascade;
create or replace function sgroups.list_svc2ag_icmp_rule()
returns setof sgroups.vu_svc2ag_icmp_rule
as $$
begin
  return query
  select * from sgroups.vu_svc2ag_icmp_rule;
end;
$$ language plpgsql stable;


drop type if exists sgroups.row_of__svc2ag_icmp_rule cascade;
create type sgroups.row_of__svc2ag_icmp_rule as (
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
  svc_local sgroups.endpoint,
  ag_remote sgroups.endpoint
);


drop function if exists sgroups.sync_svc2ag_icmp_rule(sgroups.sync_op, sgroups.row_of__svc2ag_icmp_rule) cascade;
create or replace function sgroups.sync_svc2ag_icmp_rule(
  op sgroups.sync_op, d sgroups.row_of__svc2ag_icmp_rule
)
returns setof sgroups.vu_svc2ag_icmp_rule
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
      raise exception 'sgroups.sync_svc2ag_icmp_rule: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_svc2ag_icmp_rule: namespace is required for op=ups'
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
          select * from sgroups.resolve_svc2ag_rule_targets(norm_ns, (d).svc_local, (d).ag_remote)
        )
        insert into sgroups.tbl_svc2ag_icmp_rule(
          uid, name, ns, svcLocal, agRemote, display_name, labels, annotations, comment, description, action, traffic, ip_v, entries
        )
        select
          rl_uid,
          norm_name::sgroups.rname,
          nsID,
          ids.svcl_id,
          ids.agr_id,
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
          raise exception 'svc2ag icmp rule already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_svc2ag_icmp_rule v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_svc2ag_rule_targets(norm_ns, (d).svc_local, (d).ag_remote)
      )
      update sgroups.tbl_svc2ag_icmp_rule t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             svcLocal = ids.svcl_id,
             agRemote = ids.agr_id,
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
        raise exception 'svc2ag icmp rule already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      if exists (select 1 from sgroups.tbl_svc2ag_icmp_rule where uid = (d).uid) then
        raise exception 'sgroups.sync_svc2ag_icmp_rule: update mismatch for uid=% (name or namespace differs from stored row)', (d).uid
          using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
      end if;
      if exists (select 1 from sgroups.tbl_rule_registry where uid = (d).uid) then
        raise exception 'cannot change rule type for uid=%: existing rule is of a different type', (d).uid
          using detail = 'SG0012', hint = 'rule types are immutable; delete the existing rule and create a new one to switch types';
      end if;
      raise exception 'sgroups.sync_svc2ag_icmp_rule: rule not found for uid=%', (d).uid
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_svc2ag_icmp_rule v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_svc2ag_icmp_rule: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

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

---------------------------------- RULE REFS HELPERS -------------------------------------

create index if not exists ag2ag_rule_ag_local_idx
  on sgroups.tbl_ag2ag_rule (agLocal);
create index if not exists ag2ag_rule_ag_remote_idx
  on sgroups.tbl_ag2ag_rule (agRemote);

create index if not exists ag2ag_icmp_rule_ag_local_idx
  on sgroups.tbl_ag2ag_icmp_rule (agLocal);
create index if not exists ag2ag_icmp_rule_ag_remote_idx
  on sgroups.tbl_ag2ag_icmp_rule (agRemote);

create index if not exists ag2icmp_rule_ag_local_idx
  on sgroups.tbl_ag2icmp_rule (agLocal);

create index if not exists ag2cidr_rule_ag_local_idx
  on sgroups.tbl_ag2cidr_rule (agLocal);

create index if not exists ag2cidr_icmp_rule_ag_local_idx
  on sgroups.tbl_ag2cidr_icmp_rule (agLocal);

create index if not exists ag2fqdn_rule_ag_local_idx
  on sgroups.tbl_ag2fqdn_rule (agLocal);

create index if not exists svc2svc_rule_svc_local_idx
  on sgroups.tbl_svc2svc_rule (svcLocal);
create index if not exists svc2svc_rule_svc_remote_idx
  on sgroups.tbl_svc2svc_rule (svcRemote);

create index if not exists svc2fqdn_rule_svc_local_idx
  on sgroups.tbl_svc2fqdn_rule (svcLocal);

create index if not exists svc2cidr_rule_svc_local_idx
  on sgroups.tbl_svc2cidr_rule (svcLocal);

create index if not exists svc2cidr_icmp_rule_svc_local_idx
  on sgroups.tbl_svc2cidr_icmp_rule (svcLocal);

create index if not exists ag2svc_rule_ag_local_idx
  on sgroups.tbl_ag2svc_rule (agLocal);
create index if not exists ag2svc_rule_svc_remote_idx
  on sgroups.tbl_ag2svc_rule (svcRemote);

create index if not exists svc2ag_rule_svc_local_idx
  on sgroups.tbl_svc2ag_rule (svcLocal);
create index if not exists svc2ag_rule_ag_remote_idx
  on sgroups.tbl_svc2ag_rule (agRemote);

create index if not exists svc2ag_icmp_rule_svc_local_idx
  on sgroups.tbl_svc2ag_icmp_rule (svcLocal);
create index if not exists svc2ag_icmp_rule_ag_remote_idx
  on sgroups.tbl_svc2ag_icmp_rule (agRemote);

create index if not exists ag2svc_icmp_rule_ag_local_idx
  on sgroups.tbl_ag2svc_icmp_rule (agLocal);
create index if not exists ag2svc_icmp_rule_svc_remote_idx
  on sgroups.tbl_ag2svc_icmp_rule (svcRemote);


---------------------------------- AG RULE REF HELPERS -------------------------------------

drop function if exists sgroups.get_ag2ag_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2ag_rule_refs(ag_id bigint)
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
        t.name::text,
        ns.name::text,
        'Ag2AgRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by t.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from (
    select distinct rl.id, rl.name, rl.ns
    from sgroups.tbl_ag2ag_rule rl
    where rl.agLocal = ag_id or rl.agRemote = ag_id
  ) t
  join sgroups.tbl_namespace ns on ns.id = t.ns;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2ag_rule_refs(bigint)
  is 'Returns Ag2AgRule refs where AddressGroup participates as local or remote';


drop function if exists sgroups.get_ag2ag_icmp_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2ag_icmp_rule_refs(ag_id bigint)
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
        t.name::text,
        ns.name::text,
        'Ag2AgIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by t.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from (
    select distinct rl.id, rl.name, rl.ns
    from sgroups.tbl_ag2ag_icmp_rule rl
    where rl.agLocal = ag_id or rl.agRemote = ag_id
  ) t
  join sgroups.tbl_namespace ns on ns.id = t.ns;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2ag_icmp_rule_refs(bigint)
  is 'Returns Ag2AgIcmpRule refs where AddressGroup participates as local or remote';


drop function if exists sgroups.get_ag2icmp_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2icmp_rule_refs(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Ag2IcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agLocal = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2icmp_rule_refs(bigint)
  is 'Returns Ag2IcmpRule refs where AddressGroup participates as local';


drop function if exists sgroups.get_ag2cidr_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2cidr_rule_refs(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Ag2CidrRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2cidr_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agLocal = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2cidr_rule_refs(bigint)
  is 'Returns Ag2CidrRule refs where AddressGroup participates as local';


drop function if exists sgroups.get_ag2cidr_icmp_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2cidr_icmp_rule_refs(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Ag2CidrIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2cidr_icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agLocal = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2cidr_icmp_rule_refs(bigint)
  is 'Returns Ag2CidrIcmpRule refs where AddressGroup participates as local';


drop function if exists sgroups.get_ag2fqdn_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2fqdn_rule_refs(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Ag2FqdnRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2fqdn_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agLocal = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2fqdn_rule_refs(bigint)
  is 'Returns Ag2FqdnRule refs where AddressGroup participates as local';


---------------------------------- SERVICE RULE REF HELPERS -------------------------------------

drop function if exists sgroups.get_svc2svc_rule_refs(bigint) cascade;
create or replace function sgroups.get_svc2svc_rule_refs(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        t.name::text,
        ns.name::text,
        'Svc2SvcRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by t.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from (
    select distinct rl.id, rl.name, rl.ns
    from sgroups.tbl_svc2svc_rule rl
    where rl.svcLocal = svc_id or rl.svcRemote = svc_id
  ) t
  join sgroups.tbl_namespace ns on ns.id = t.ns;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2svc_rule_refs(bigint)
  is 'Returns Svc2SvcRule refs where Service participates as local or remote';


drop function if exists sgroups.get_svc2fqdn_rule_refs(bigint) cascade;
create or replace function sgroups.get_svc2fqdn_rule_refs(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Svc2FqdnRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2fqdn_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcLocal = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2fqdn_rule_refs(bigint)
  is 'Returns Svc2FqdnRule refs where Service participates as local';


drop function if exists sgroups.get_svc2cidr_rule_refs(bigint) cascade;
create or replace function sgroups.get_svc2cidr_rule_refs(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Svc2CidrRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2cidr_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcLocal = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2cidr_rule_refs(bigint)
  is 'Returns Svc2CidrRule refs where Service participates as local';


drop function if exists sgroups.get_svc2cidr_icmp_rule_refs(bigint) cascade;
create or replace function sgroups.get_svc2cidr_icmp_rule_refs(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Svc2CidrIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2cidr_icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcLocal = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2cidr_icmp_rule_refs(bigint)
  is 'Returns Svc2CidrIcmpRule refs where Service participates as local';


---------------------------------- ASYMMETRIC AG↔SVC RULE REF HELPERS ----------------------

drop function if exists sgroups.get_ag2svc_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2svc_rule_refs(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Ag2SvcRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2svc_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agLocal = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2svc_rule_refs(bigint)
  is 'Returns Ag2SvcRule refs where AddressGroup participates as local';


drop function if exists sgroups.get_ag2svc_rule_refs_by_svc(bigint) cascade;
create or replace function sgroups.get_ag2svc_rule_refs_by_svc(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Ag2SvcRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2svc_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcRemote = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2svc_rule_refs_by_svc(bigint)
  is 'Returns Ag2SvcRule refs where Service participates as remote';


drop function if exists sgroups.get_svc2ag_rule_refs(bigint) cascade;
create or replace function sgroups.get_svc2ag_rule_refs(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Svc2AgRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2ag_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcLocal = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2ag_rule_refs(bigint)
  is 'Returns Svc2AgRule refs where Service participates as local';


drop function if exists sgroups.get_svc2ag_rule_refs_by_ag(bigint) cascade;
create or replace function sgroups.get_svc2ag_rule_refs_by_ag(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Svc2AgRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2ag_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agRemote = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2ag_rule_refs_by_ag(bigint)
  is 'Returns Svc2AgRule refs where AddressGroup participates as remote';


drop function if exists sgroups.get_svc2ag_icmp_rule_refs(bigint) cascade;
create or replace function sgroups.get_svc2ag_icmp_rule_refs(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Svc2AgIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2ag_icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcLocal = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2ag_icmp_rule_refs(bigint)
  is 'Returns Svc2AgIcmpRule refs where Service participates as local';


drop function if exists sgroups.get_svc2ag_icmp_rule_refs_by_ag(bigint) cascade;
create or replace function sgroups.get_svc2ag_icmp_rule_refs_by_ag(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Svc2AgIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_svc2ag_icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agRemote = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_svc2ag_icmp_rule_refs_by_ag(bigint)
  is 'Returns Svc2AgIcmpRule refs where AddressGroup participates as remote';


drop function if exists sgroups.get_ag2svc_icmp_rule_refs(bigint) cascade;
create or replace function sgroups.get_ag2svc_icmp_rule_refs(ag_id bigint)
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
        rl.name::text,
        ns.name::text,
        'Ag2SvcIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2svc_icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.agLocal = ag_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2svc_icmp_rule_refs(bigint)
  is 'Returns Ag2SvcIcmpRule refs where AddressGroup participates as local';


drop function if exists sgroups.get_ag2svc_icmp_rule_refs_by_svc(bigint) cascade;
create or replace function sgroups.get_ag2svc_icmp_rule_refs_by_svc(svc_id bigint)
returns sgroups.resource_ref[]
as $$
declare
  r sgroups.resource_ref[];
begin
  if svc_id is null then
    return array[]::sgroups.resource_ref[];
  end if;

  select coalesce(
    array_agg(
      row(
        rl.name::text,
        ns.name::text,
        'Ag2SvcIcmpRule'::sgroups.resource_type
      )::sgroups.resource_ref
      order by rl.id
    ),
    array[]::sgroups.resource_ref[]
  )
  into r
  from sgroups.tbl_ag2svc_icmp_rule rl
  join sgroups.tbl_namespace ns on ns.id = rl.ns
  where rl.svcRemote = svc_id;

  return r;
end;
$$ language plpgsql stable;

comment on function sgroups.get_ag2svc_icmp_rule_refs_by_svc(bigint)
  is 'Returns Ag2SvcIcmpRule refs where Service participates as remote';


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
  sgroups.get_host_refs(ag.id)
    || sgroups.get_network_refs(ag.id)
    || sgroups.get_service_refs(ag.id)
    || sgroups.get_ag2ag_rule_refs(ag.id)
    || sgroups.get_ag2ag_icmp_rule_refs(ag.id)
    || sgroups.get_ag2icmp_rule_refs(ag.id)
    || sgroups.get_ag2cidr_rule_refs(ag.id)
    || sgroups.get_ag2cidr_icmp_rule_refs(ag.id)
    || sgroups.get_ag2fqdn_rule_refs(ag.id)
    || sgroups.get_ag2svc_rule_refs(ag.id)
    || sgroups.get_ag2svc_icmp_rule_refs(ag.id)
    || sgroups.get_svc2ag_rule_refs_by_ag(ag.id)
    || sgroups.get_svc2ag_icmp_rule_refs_by_ag(ag.id) as refs,
  ag.creation_timestamp,
  ag.resource_version
from sgroups.tbl_ag as ag;

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
  )
    || sgroups.get_svc2svc_rule_refs(svc.id)
    || sgroups.get_svc2fqdn_rule_refs(svc.id)
    || sgroups.get_svc2cidr_rule_refs(svc.id)
    || sgroups.get_svc2cidr_icmp_rule_refs(svc.id)
    || sgroups.get_ag2svc_rule_refs_by_svc(svc.id)
    || sgroups.get_ag2svc_icmp_rule_refs_by_svc(svc.id)
    || sgroups.get_svc2ag_rule_refs(svc.id)
    || sgroups.get_svc2ag_icmp_rule_refs(svc.id) as refs,
  svc.creation_timestamp,
  svc.resource_version
from sgroups.tbl_service as svc
left join sgroups.tbl_namespace ns on ns.id = svc.ns;

drop trigger if exists trg_upd_ag2ag_rule_refs on sgroups.tbl_ag2ag_rule;
create trigger trg_upd_ag2ag_rule_refs
after insert or update or delete on sgroups.tbl_ag2ag_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal',
  'sgroups.tbl_ag', 'agremote'
);

drop trigger if exists trg_upd_ag2ag_icmp_rule_refs on sgroups.tbl_ag2ag_icmp_rule;
create trigger trg_upd_ag2ag_icmp_rule_refs
after insert or update or delete on sgroups.tbl_ag2ag_icmp_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal',
  'sgroups.tbl_ag', 'agremote'
);

drop trigger if exists trg_upd_ag2icmp_rule_refs on sgroups.tbl_ag2icmp_rule;
create trigger trg_upd_ag2icmp_rule_refs
after insert or update or delete on sgroups.tbl_ag2icmp_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal'
);

drop trigger if exists trg_upd_ag2cidr_rule_refs on sgroups.tbl_ag2cidr_rule;
create trigger trg_upd_ag2cidr_rule_refs
after insert or update or delete on sgroups.tbl_ag2cidr_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal'
);

drop trigger if exists trg_upd_ag2cidr_icmp_rule_refs on sgroups.tbl_ag2cidr_icmp_rule;
create trigger trg_upd_ag2cidr_icmp_rule_refs
after insert or update or delete on sgroups.tbl_ag2cidr_icmp_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal'
);

drop trigger if exists trg_upd_ag2fqdn_rule_refs on sgroups.tbl_ag2fqdn_rule;
create trigger trg_upd_ag2fqdn_rule_refs
after insert or update or delete on sgroups.tbl_ag2fqdn_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal'
);

drop trigger if exists trg_upd_svc2svc_rule_refs on sgroups.tbl_svc2svc_rule;
create trigger trg_upd_svc2svc_rule_refs
after insert or update or delete on sgroups.tbl_svc2svc_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_service', 'svclocal',
  'sgroups.tbl_service', 'svcremote'
);

drop trigger if exists trg_upd_svc2fqdn_rule_refs on sgroups.tbl_svc2fqdn_rule;
create trigger trg_upd_svc2fqdn_rule_refs
after insert or update or delete on sgroups.tbl_svc2fqdn_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_service', 'svclocal'
);

drop trigger if exists trg_upd_svc2cidr_rule_refs on sgroups.tbl_svc2cidr_rule;
create trigger trg_upd_svc2cidr_rule_refs
after insert or update or delete on sgroups.tbl_svc2cidr_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_service', 'svclocal'
);

drop trigger if exists trg_upd_svc2cidr_icmp_rule_refs on sgroups.tbl_svc2cidr_icmp_rule;
create trigger trg_upd_svc2cidr_icmp_rule_refs
after insert or update or delete on sgroups.tbl_svc2cidr_icmp_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_service', 'svclocal'
);

drop trigger if exists trg_upd_ag2svc_rule_refs on sgroups.tbl_ag2svc_rule;
create trigger trg_upd_ag2svc_rule_refs
after insert or update or delete on sgroups.tbl_ag2svc_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal',
  'sgroups.tbl_service', 'svcremote'
);

drop trigger if exists trg_upd_svc2ag_rule_refs on sgroups.tbl_svc2ag_rule;
create trigger trg_upd_svc2ag_rule_refs
after insert or update or delete on sgroups.tbl_svc2ag_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_service', 'svclocal',
  'sgroups.tbl_ag', 'agremote'
);

drop trigger if exists trg_upd_svc2ag_icmp_rule_refs on sgroups.tbl_svc2ag_icmp_rule;
create trigger trg_upd_svc2ag_icmp_rule_refs
after insert or update or delete on sgroups.tbl_svc2ag_icmp_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_service', 'svclocal',
  'sgroups.tbl_ag', 'agremote'
);

drop trigger if exists trg_upd_ag2svc_icmp_rule_refs on sgroups.tbl_ag2svc_icmp_rule;
create trigger trg_upd_ag2svc_icmp_rule_refs
after insert or update or delete on sgroups.tbl_ag2svc_icmp_rule
for each row
execute function sgroups.upd_refs_trg(
  'rule_rev',
  'sgroups.tbl_ag', 'aglocal',
  'sgroups.tbl_service', 'svcremote'
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
