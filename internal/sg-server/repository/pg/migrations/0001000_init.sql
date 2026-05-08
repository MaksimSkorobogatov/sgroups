-- +goose Up
-- +goose StatementBegin

create schema if not exists sgroups;

create extension if not exists btree_gist;
create extension if not exists hstore;
create extension if not exists pgcrypto;

--------------------------------------------- AGGREGATES ----------------------------------------------

drop type if exists sgroups.i4mr_any_intersect_state cascade;
create type sgroups.i4mr_any_intersect_state as (
    overlapped boolean,
    state int4multirange,
    nulls int
);

drop function if exists sgroups.i4mr_any_intersect_s(sgroups.i4mr_any_intersect_state, int4multirange) cascade;
create or replace function sgroups.i4mr_any_intersect_s(
    s sgroups.i4mr_any_intersect_state,
    val int4multirange
)
returns sgroups.i4mr_any_intersect_state
as $$
    select case
             when (s).overlapped
               then s
             when val is null
               then row((s).overlapped, (s).state, (s).nulls + 1)::sgroups.i4mr_any_intersect_state
             when (s).state && val
               then row(true, (s).state * val, (s).nulls)::sgroups.i4mr_any_intersect_state
             else
               row((s).overlapped, (s).state + val, (s).nulls)::sgroups.i4mr_any_intersect_state
           end;
$$ language sql immutable;

drop function if exists sgroups.i4mr_any_intersect_f(sgroups.i4mr_any_intersect_state) cascade;
create or replace function sgroups.i4mr_any_intersect_f(v sgroups.i4mr_any_intersect_state)
returns boolean
as $$
    select
      (v).overlapped
      or (v).nulls > 1
      or ((v).nulls > 0 and lower((v).state) is not null);
$$ language sql immutable strict;

drop aggregate if exists sgroups.i4mr_any_intersect(int4multirange) cascade;
create aggregate sgroups.i4mr_any_intersect(int4multirange) (
    stype = sgroups.i4mr_any_intersect_state,
    sfunc = sgroups.i4mr_any_intersect_s,
    finalfunc = sgroups.i4mr_any_intersect_f,
    initcond = '(false, "{}", 0)'
);

comment on aggregate sgroups.i4mr_any_intersect(int4multirange)
  is 'Detects intersection among int4multirange set; treats NULLs as intersecting if mixed with non-NULL or if more than one NULL';

---------------------------------- COMMON TYPES & DOMAINS -------------------------------------

drop type if exists sgroups.policy_action cascade;
create type sgroups.policy_action as enum (
    'DENY',
    'ALLOW'
);

drop type if exists sgroups.proto cascade;
create type sgroups.proto as enum (
   'tcp',
   'udp',
   'icmp'
);
comment on type sgroups.proto is 'transport protocol in IP net';

drop domain if exists sgroups.port_ranges cascade;
create domain sgroups.port_ranges
           as int4multirange
   constraint port_ranges_correctness
        check (
            value <@ '[1, 65536)'::int4range
        )
   constraint port_ranges_emptiness
        check (
            not (value = '{}')
        );
comment on domain sgroups.port_ranges
     is 'port ranges used by Security Group Rule';

drop domain if exists sgroups.sync_op cascade;
create domain sgroups.sync_op
           as varchar(3)
    constraint sync_op_value
        check (
            value is not null and
            value = any (array['ups', 'del'])
        );
comment on domain sgroups.sync_op is 'type operation of UPSERT, DELETE';


drop type if exists sgroups.resource_type cascade;
create type sgroups.resource_type as enum (
  'Namespace',
  'AddressGroup',
  'Host',
  'HostBinding',
  'Service',
  'ServiceBinding',
  'Network',
  'NetworkBinding',
  'Ag2AgRule',
  'Ag2AgIcmpRule',
  'Ag2IcmpRule',
  'Ag2CidrRule',
  'Ag2CidrIcmpRule',
  'Ag2FqdnRule',
  'Ag2SvcRule',
  'Ag2SvcIcmpRule',
  'Svc2SvcRule',
  'Svc2FqdnRule',
  'Svc2CidrRule',
  'Svc2CidrIcmpRule',
  'Svc2AgRule',
  'Svc2AgIcmpRule'
);
comment on type sgroups.resource_type
  is 'type of resource (Namespace, AddressGroup, Host, HostBinding, Service, ServiceBinding, Network, NetworkBinding, Ag2AgRule, Ag2AgIcmpRule, Ag2IcmpRule, Ag2CidrRule, Ag2CidrIcmpRule, Ag2FqdnRule, Svc2SvcRule, Svc2FqdnRule)';


drop domain if exists sgroups.resource_op cascade;
create domain sgroups.resource_op
           as text
   constraint resource_op_value
        check (
            value is not null and
            value = any (array['ADDED','MODIFIED','DELETED'])
        );
comment on domain sgroups.resource_op is 'type resource operation of ADD, MODIFY, DELETE';

drop domain if exists sgroups.rname cascade;
create domain sgroups.rname
           as text
   constraint rname_length
        check (
            length(value) BETWEEN 1 AND 63
        )
   constraint rname_validity
        check (
            (value ~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$')
        );
comment on domain sgroups.rname is 'Resource Names type';

drop domain if exists sgroups.dname cascade;
create domain sgroups.dname
           as text
   constraint dname_length
        check (
            length(value) BETWEEN 0 AND 255
        );
comment on domain sgroups.dname is 'Display Names type';


drop type if exists sgroups.resource_ref cascade;
create type sgroups.resource_ref as (
  name      text,
  namespace text,
  res_type  sgroups.resource_type
);
comment on type sgroups.resource_ref is 'resource reference type';

drop type if exists sgroups.field_selector cascade;
create type sgroups.field_selector as (
  name      text,
  namespace text,
  refs      sgroups.resource_ref[]
);
comment on type sgroups.field_selector is 'resource field selector type';

drop type if exists sgroups.res_selector cascade;
create type sgroups.res_selector as (
  field_selector sgroups.field_selector,
  label_selector hstore
);
comment on type sgroups.res_selector is 'resource selector type: field_selector for equality-based matching on name/namespace/refs, label_selector for subset matching on labels';


drop type if exists sgroups.resource_id cascade;
create type sgroups.resource_id as (
  name      text,
  namespace text
);
comment on type sgroups.resource_id is 'resource identifier type';


drop type if exists sgroups.traffic cascade;
create type sgroups.traffic as enum (
    'both',
    'ingress',
    'egress'
);

drop type if exists sgroups.ip_family cascade;
create type sgroups.ip_family as enum (
    'IPv4',
    'IPv6'
);

---------------------------------- COMMON TABLES ----------------------------------------
drop table if exists sgroups.tbl_global_resource_version cascade;
create table if not exists sgroups.tbl_global_resource_version (
  id smallint primary key,
  v  bigint not null check (v >= 0)
);
comment on table sgroups.tbl_global_resource_version is 'Singleton table for global resourceVersion counter (id=1)';

insert into sgroups.tbl_global_resource_version (id, v)
values (1, 0)
on conflict (id) do nothing;

drop table if exists sgroups.tbl_outbox_resource_events cascade;
create table if not exists sgroups.tbl_outbox_resource_events (
  event_id        bigint generated always as identity primary key,
  ts              timestamptz not null default clock_timestamp(),
  resource_version bigint not null,
  resource_type   sgroups.resource_type not null,
  event_type      sgroups.resource_op,
  object          jsonb not null
);
comment on table sgroups.tbl_outbox_resource_events is 'Outbox table for resource events stream';

create index if not exists outbox_resource_events_type_rv_event_id_idx
  on sgroups.tbl_outbox_resource_events (resource_type, resource_version, event_id);

create index if not exists outbox_resource_events_ts_idx
  on sgroups.tbl_outbox_resource_events (ts);

---------------------------------- COMMON FUNCTIONS AND TRIGGERS-------------------------------------

drop function if exists sgroups.ts() cascade;
create or replace function sgroups.ts()
    returns timestamptz
as $$
begin
    return now();
end;
$$ language plpgsql immutable;
comment on function sgroups.ts
     is 'this is simple "Timestamp" function but with "immutable" flag';


drop function if exists sgroups.ports_dont_intersect(sgroups.port_ranges[]);
create or replace function sgroups.ports_dont_intersect(src sgroups.port_ranges[])
    returns boolean
as $$
    declare
        ret boolean := false;
    begin
        with items as (
           select unnest(src) as value
        ) select sgroups.i4mr_any_intersect(value)
            from items
            into ret;
        return not ret;
    end;
$$ language plpgsql;
comment on function sgroups.ports_dont_intersect
     is 'check ports array has not intersections';


drop function if exists sgroups.mk_field_selector(text, text, sgroups.resource_ref[]) cascade;
create or replace function sgroups.mk_field_selector(
  name text,
  namespace text,
  refs sgroups.resource_ref[]
)
returns sgroups.field_selector
as $$
begin
  return row(
    nullif(btrim(name), ''),
    nullif(btrim(namespace), ''),
    refs
  )::sgroups.field_selector;
end;
$$ language plpgsql immutable;
comment on function sgroups.mk_field_selector(text, text, sgroups.resource_ref[])
  is 'Constructor helper: builds field_selector (name, namespace, refs)';

drop function if exists sgroups.mk_res_selector(text, text, sgroups.resource_ref[], hstore) cascade;
create or replace function sgroups.mk_res_selector(
  name text,
  namespace text,
  refs sgroups.resource_ref[],
  labels hstore
)
returns sgroups.res_selector
as $$
begin
  return row(
    sgroups.mk_field_selector(name, namespace, refs),
    labels
  )::sgroups.res_selector;
end;
$$ language plpgsql immutable;
comment on function sgroups.mk_res_selector(text, text, sgroups.resource_ref[], hstore)
  is 'Constructor helper: builds res_selector (field_selector + label_selector)';


drop function if exists sgroups.get_resource_version() cascade;
create or replace function sgroups.get_resource_version()
returns text
as $$
  select coalesce(
    (select v::text from sgroups.tbl_global_resource_version where id = 1),
    '0'
  );
$$ language sql stable;
comment on function sgroups.get_resource_version() is 'Returns current global resourceVersion as text';


drop function if exists sgroups.next_resource_version() cascade;
create or replace function sgroups.next_resource_version()
returns bigint
as $$
declare
  new_v bigint;
begin
  insert into sgroups.tbl_global_resource_version (id, v)
  values (1, 1)
  on conflict (id) do update
    set v = sgroups.tbl_global_resource_version.v + 1
  returning v into new_v;

  return new_v;
end;
$$ language plpgsql;
comment on function sgroups.next_resource_version() is 'Increments and returns global resourceVersion';


drop function if exists sgroups.is_empty_uuid(uuid) cascade;
create or replace function sgroups.is_empty_uuid(u uuid)
returns boolean
as $$
  select u is null or u = '00000000-0000-0000-0000-000000000000'::uuid;
$$ language sql immutable;
comment on function sgroups.is_empty_uuid(uuid)
  is 'True when uuid is NULL or zero-uuid';


drop function if exists sgroups.match_res_selector(sgroups.res_selector, sgroups.res_selector) cascade;
create or replace function sgroups.match_res_selector(
  res sgroups.res_selector,
  sel sgroups.res_selector
)
returns boolean
as $$
declare
  empty_labels hstore;
  sel_name text;
  sel_ns text;
begin
  empty_labels := hstore(array[]::text[], array[]::text[]);

  -- Treat empty strings as NULL for selector semantics.
  sel_name := nullif(btrim((sel.field_selector).name), '');
  sel_ns := nullif(btrim((sel.field_selector).namespace), '');

  return
    (
      (sel_name is null)
      or nullif(btrim((res.field_selector).name), '') = sel_name
    )
    and
    (
      (sel_ns is null)
      or nullif(btrim((res.field_selector).namespace), '') = sel_ns
    )
    and
    (
      sel.label_selector is null
      or coalesce(res.label_selector, empty_labels) @> sel.label_selector
    )
    and
    (
      -- selector.refs: missing/empty => no filtering
      (sel.field_selector).refs is null
      or coalesce(cardinality((sel.field_selector).refs), 0) = 0
      or (
        -- otherwise: every selector ref must have at least one partial-matching resource ref (AND across selector refs)
        (res.field_selector).refs is not null
        and coalesce(cardinality((res.field_selector).refs), 0) > 0
        and not exists (
          select 1
          from unnest((sel.field_selector).refs) sr
          where not exists (
            select 1
            from unnest((res.field_selector).refs) rr
            where (sr.res_type is null or sr.res_type = rr.res_type)
              and (nullif(btrim(sr.name), '') is null or nullif(btrim(rr.name), '') = nullif(btrim(sr.name), ''))
              and (nullif(btrim(sr.namespace), '') is null or nullif(btrim(rr.namespace), '') = nullif(btrim(sr.namespace), ''))
          )
        )
      )
    );
end;
$$ language plpgsql immutable;
comment on function sgroups.match_res_selector(sgroups.res_selector, sgroups.res_selector)
  is 'res_selector match predicate: AND inside selector (name/namespace/labels/refs). Every selector ref must partial-match at least one resource ref';


drop operator if exists public.~> (sgroups.res_selector, sgroups.res_selector);
create operator public.~> (
  procedure = sgroups.match_res_selector,
  leftarg   = sgroups.res_selector,
  rightarg  = sgroups.res_selector
);
comment on operator public.~>(sgroups.res_selector, sgroups.res_selector)
  is 'Match operator: res ~> sel means "resource matches selector"';


drop function if exists sgroups.match_res_selectors(sgroups.res_selector, sgroups.res_selector[]) cascade;
create or replace function sgroups.match_res_selectors(
  res sgroups.res_selector,
  selectors sgroups.res_selector[] default null
)
returns boolean
as $$
begin
  if selectors is null or cardinality(selectors) = 0 then
    return true;
  end if;

  return res ~> ANY (selectors);
end;
$$ language plpgsql immutable;
comment on function sgroups.match_res_selectors(sgroups.res_selector, sgroups.res_selector[])
  is 'OR across selectors array using ~> match; empty selectors => match all';


drop function if exists sgroups.del_expired_outbox_resource_events(interval) cascade;
create or replace function sgroups.del_expired_outbox_resource_events(
  older_than interval default interval '5 minutes'
)
returns integer
as $$
declare
  deleted_count integer;
begin
  with doomed as (
    select event_id
    from sgroups.tbl_outbox_resource_events
    where ts < now() - older_than
    order by ts
  )
  delete from sgroups.tbl_outbox_resource_events e
  using doomed d
  where e.event_id = d.event_id;

  get diagnostics deleted_count = row_count;
  return deleted_count;
end;
$$ language plpgsql;
comment on function sgroups.del_expired_outbox_resource_events(interval) is 'Deletes outbox events older than retention window (best run periodically)';


drop function if exists sgroups.inc_resource_version_cnt() cascade;
create or replace function sgroups.inc_resource_version_cnt()
returns trigger
as $$
declare
  rv bigint;
begin
  if TG_OP <> 'INSERT' and TG_OP <> 'UPDATE' then
    raise exception 'sgroups.inc_resource_version_cnt: unsupported TG_OP=%', TG_OP
      using detail = 'SG0003', hint = 'trigger must be fired only for INSERT/UPDATE';
  end if;

  rv := sgroups.next_resource_version();

  NEW.resource_version := rv::text;
  return NEW;
end;
$$ language plpgsql;
comment on function sgroups.inc_resource_version_cnt()
  is 'BEFORE INSERT/UPDATE trigger: increments global resourceVersion and sets NEW.resource_version';


drop function if exists sgroups.capture_outbox_resource_events_trg() cascade;
create or replace function sgroups.capture_outbox_resource_events_trg()
returns trigger
as $$
declare
  rv bigint;
  payload jsonb;
  evt_type text;
  r_type sgroups.resource_type;
  view_name regclass;
  target_uid uuid;
begin
  if TG_NARGS < 2 or TG_ARGV[0] is null or TG_ARGV[0] = '' or TG_ARGV[1] is null or TG_ARGV[1] = '' then
    raise exception
      'sgroups.capture_outbox_resource_events_trg: missing args (resource_type, view); table %.%, trigger %',
      TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_NAME
      using detail = 'SG0003', hint = 'pass resource_type and schema-qualified view name';
  end if;

  r_type := TG_ARGV[0]::sgroups.resource_type;
  view_name := TG_ARGV[1]::regclass;

  case TG_OP
    when 'INSERT' then
      evt_type := 'ADDED';
    when 'UPDATE' then
      evt_type := 'MODIFIED';
    when 'DELETE' then
      evt_type := 'DELETED';
    else
      raise exception 'sgroups.capture_outbox_resource_events_trg: unsupported TG_OP=%', TG_OP
        using detail = 'SG0003', hint = 'trigger must be fired only for INSERT/UPDATE/DELETE';
  end case;

  if TG_OP = 'DELETE' then
    -- Must run BEFORE DELETE to be able to read payload from the view.
    if TG_WHEN <> 'BEFORE' then
      raise exception 'sgroups.capture_outbox_resource_events_trg: DELETE must run in BEFORE trigger; table %.%, trigger %',
        TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_NAME
        using detail = 'SG0003', hint = 'attach this trigger as BEFORE DELETE';
    end if;

    rv := sgroups.next_resource_version();
    target_uid := OLD.uid;
  else
    -- Must run AFTER INSERT/UPDATE to read a fully-populated view representation.
    if TG_WHEN <> 'AFTER' then
      raise exception 'sgroups.capture_outbox_resource_events_trg: % must run in AFTER trigger; table %.%, trigger %',
        TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_NAME
        using detail = 'SG0003', hint = 'attach this trigger as AFTER INSERT OR UPDATE';
    end if;

    if NEW.resource_version is null or btrim(NEW.resource_version) = '' then
      raise exception 'sgroups.capture_outbox_resource_events_trg: NEW.resource_version is required for % on %.%',
        TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME
        using detail = 'SG0004', hint = 'attach inc_resource_version_cnt() as BEFORE INSERT OR UPDATE';
    end if;
    rv := NEW.resource_version::bigint;
    target_uid := NEW.uid;
  end if;

  execute format('select to_jsonb(v) from %s v where v.uid = $1', view_name)
    into payload
    using target_uid;

  if payload is null then
    raise exception 'sgroups.capture_outbox_resource_events_trg: cannot read payload from view=% for uid=% (table %.%, trigger %)',
      view_name::text, target_uid, TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_NAME
      using detail = 'SG0004', hint = 'ensure the view exposes uid and returns the row for this uid';
  end if;

  -- Never leak internal surrogate PKs via outbox / LISTEN-NOTIFY payloads.
  -- Safe no-op if the key doesn't exist.
  payload := payload - 'id';

  insert into sgroups.tbl_outbox_resource_events (
    resource_version,
    resource_type,
    event_type,
    object
  ) values (
    rv,
    r_type,
    evt_type::sgroups.resource_op,
    payload
  );

  -- Retention window (LRU-ish): keep only recent history.
  perform sgroups.del_expired_outbox_resource_events(interval '5 minutes');

  if TG_OP = 'DELETE' then
    return OLD;
  end if;
  return NEW;
end;
$$ language plpgsql;
comment on function sgroups.capture_outbox_resource_events_trg() 
  is 'Writes outbox resource event using payload from view(arg#2). For INSERT/UPDATE runs AFTER and uses NEW.resource_version; for DELETE runs BEFORE and allocates a new resource_version.';


drop function if exists sgroups.notify_outbox_resource_events_trg() cascade;
create or replace function sgroups.notify_outbox_resource_events_trg()
returns trigger
as $$
declare
  notify_channel text;
  notify_payload jsonb;
begin
  notify_channel := format('resource_%s', NEW.resource_type::text);
  notify_payload := jsonb_build_object(
    'ts', NEW.ts,
    'resource_version', NEW.resource_version::text,
    'resource_type', NEW.resource_type::text,
    'event_type', NEW.event_type::text,
    'object', NEW.object
  );

  begin
    perform pg_notify(notify_channel, notify_payload::text);
    exception when others then
    -- Best-effort notification: never block/abort the transaction that writes data.
    null;
  end;
  return NEW;
end;
$$ language plpgsql;
comment on function sgroups.notify_outbox_resource_events_trg() is 'Sends NOTIFY for outbox events (payload matches list_outbox_resource_events output)';


drop trigger if exists trg_outbox_resource_events_notify on sgroups.tbl_outbox_resource_events;
create trigger trg_outbox_resource_events_notify
after insert on sgroups.tbl_outbox_resource_events
for each row
execute function sgroups.notify_outbox_resource_events_trg();


drop function if exists sgroups.list_outbox_resource_events(sgroups.resource_type, bigint, integer) cascade;
create or replace function sgroups.list_outbox_resource_events(
  r_type sgroups.resource_type default null,
  since_rv bigint default null,
  max_events integer default 10000
)
returns table (
  ts timestamptz,
  resource_version text,
  resource_type sgroups.resource_type,
  event_type sgroups.resource_op,
  object jsonb
)
as $$
declare
  cur_rv bigint;
  min_rv bigint;
begin
  if since_rv is not null then
    select v into cur_rv
    from sgroups.tbl_global_resource_version
    where id = 1;

    cur_rv := coalesce(cur_rv, 0);

    if since_rv > cur_rv then
      raise exception 'resourceVersion not found: since=% current=%', since_rv, cur_rv
        using detail = 'SG0002', hint = 'resourceVersion not found';
    end if;

    -- Global expired check: use min(resource_version) across ALL types,
    -- because resource_version is a global counter shared by all resource types.
    select min(e.resource_version) into min_rv
    from sgroups.tbl_outbox_resource_events e;

    if min_rv is null then
      -- Outbox is completely empty (all events expired or none ever created).
      if since_rv < cur_rv then
        raise exception 'resourceVersion expired (outbox empty, requested type=%): since=% current=%', r_type, since_rv, cur_rv
          using detail = 'SG0001', hint = 'resourceVersion expired';
      end if;
      return;
    end if;

    if since_rv < (min_rv - 1) then
      raise exception 'resourceVersion expired (requested type=%): since=% min_retained=% current=%', r_type, since_rv, min_rv, cur_rv
        using detail = 'SG0001', hint = 'resourceVersion expired';
    end if;
  end if;

  return query
  select
    e.ts,
    e.resource_version::text,
    e.resource_type,
    e.event_type,
    e.object
  from sgroups.tbl_outbox_resource_events e
  where (r_type is null or e.resource_type = r_type)
    and (since_rv is null or e.resource_version > since_rv)
  order by e.resource_version, e.event_id
  limit greatest(0, coalesce(max_events, 0));
end;
$$ language plpgsql stable;
comment on function sgroups.list_outbox_resource_events(sgroups.resource_type, bigint, integer) is 'Returns outbox resource events after since_rv; raises P0001 expired or P0002 not_found';


drop function if exists sgroups.forbid_fields_update_trg() cascade;
create or replace function sgroups.forbid_fields_update_trg()
returns trigger
as $$
declare
  oldj jsonb;
  newj jsonb;
  raw text;
  k text;
  ks text[];
begin
  -- Universal immutability guard: forbids changing fields listed in trigger args.
  -- Usage: execute function sgroups.forbid_fields_update_trg('uid','name'[, ...])
  -- Also accepts comma-separated lists inside an arg: ('uid,name').
  if TG_NARGS < 1 then
    raise exception
      'sgroups.forbid_fields_update_trg: missing immutable fields args; table %.%, trigger %',
      TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_NAME
      using detail = 'SG0006', hint = 'pass immutable field names as trigger args';
  end if;

  oldj := to_jsonb(OLD);
  newj := to_jsonb(NEW);

  ks := array[]::text[];
  for i in 0..TG_NARGS-1 loop
    raw := TG_ARGV[i];
    if raw is null or btrim(raw) = '' then
      continue;
    end if;
    ks := ks || regexp_split_to_array(raw, '\s*,\s*');
  end loop;

  ks := array_remove(ks, null);
  ks := array_remove(ks, '');

  if ks is null or cardinality(ks) = 0 then
    raise exception
      'sgroups.forbid_fields_update_trg: empty immutable fields args; table %.%, trigger %',
      TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_NAME
      using detail = 'SG0006', hint = 'pass immutable field names as trigger args';
  end if;

  foreach k in array ks loop
    -- Only validate if the column exists in the target table.
    if (newj ? k) and (oldj ? k) and ((newj -> k) is distinct from (oldj -> k)) then
      raise exception '%I.%I is immutable', TG_TABLE_NAME, k
        using detail = 'SG0005', hint = format('%s cannot be updated', k);
    end if;
  end loop;

  return NEW;
end;
$$ language plpgsql;
comment on function sgroups.forbid_fields_update_trg()
  is 'BEFORE UPDATE trigger: prevents changing fields listed in trigger args (if columns exist)';


drop function if exists sgroups.upd_refs_trg() cascade;
create or replace function sgroups.upd_refs_trg()
returns trigger
as $$
declare
  rev_col text;
  i int;
  parent_rel regclass;
  ref_col text;
  old_id bigint;
  new_id bigint;
begin
  if tg_op <> 'INSERT' and tg_op <> 'UPDATE' and tg_op <> 'DELETE' then
    raise exception 'sgroups.upd_refs_trg: unsupported TG_OP=%', tg_op
      using detail = 'SG0003', hint = 'trigger must be fired only for INSERT/UPDATE/DELETE';
  end if;

  if tg_nargs is null or tg_nargs < 1 then
    raise exception 'sgroups.upd_refs_trg: expected at least 1 arg (rev_column_name), got %', coalesce(tg_nargs, 0)
      using detail = 'SG0007', hint = 'pass (rev_column_name, [parent_table, ref_column]*)';
  end if;

  rev_col := tg_argv[0];

  if tg_nargs = 1 then
    if tg_op = 'DELETE' then
      return old;
    end if;
    return new;
  end if;

  if tg_nargs % 2 = 0 then
    raise exception 'sgroups.upd_refs_trg: expected (rev_column_name, then pairs of parent_table, ref_column), got % args', tg_nargs
      using detail = 'SG0007', hint = 'pass (rev_column_name, then pairs: regclass, column_name, regclass, column_name, ...)';
  end if;

  i := 1;
  while i < tg_nargs loop
    parent_rel := tg_argv[i]::regclass;
    ref_col := tg_argv[i+1];

    old_id := null;
    new_id := null;

    if tg_op = 'INSERT' then
      new_id := nullif(to_jsonb(new)->>ref_col, '')::bigint;
      if new_id is not null then
        execute format('update %s set %I = %I + 1 where id = $1', parent_rel, rev_col, rev_col)
          using new_id;
      end if;

    elsif tg_op = 'DELETE' then
      old_id := nullif(to_jsonb(old)->>ref_col, '')::bigint;
      if old_id is not null then
        execute format('update %s set %I = %I + 1 where id = $1', parent_rel, rev_col, rev_col)
          using old_id;
      end if;

    else -- UPDATE
      old_id := nullif(to_jsonb(old)->>ref_col, '')::bigint;
      new_id := nullif(to_jsonb(new)->>ref_col, '')::bigint;

      if new_id is distinct from old_id then
        if old_id is not null then
          execute format('update %s set %I = %I + 1 where id = $1', parent_rel, rev_col, rev_col)
            using old_id;
        end if;
        if new_id is not null then
          execute format('update %s set %I = %I + 1 where id = $1', parent_rel, rev_col, rev_col)
            using new_id;
        end if;
      end if;
    end if;

    i := i + 2;
  end loop;

  if tg_op = 'DELETE' then
    return old;
  end if;
  return new;
end;
$$ language plpgsql;
comment on function sgroups.upd_refs_trg()
  is 'generic trigger: bumps the `*_rev` column named by first tg_argv on each parent table passed as subsequent (regclass, ref_column) pairs';

drop function if exists sgroups.icmp_type_values_ok(int2[]) cascade;
create or replace function sgroups.icmp_type_values_ok(vals int2[])
   returns boolean
as $$
declare badVal int2;
begin
   select c
     from unnest(vals) as c
    where not (c between 0 and 255)
    limit 1
     into badVal;
   return badVal is null;
end;
$$ language plpgsql immutable strict;

drop domain if exists sgroups.icmp_types cascade;
create domain sgroups.icmp_types
           as int2[]
   constraint validate_icmp_type_values
        check (
            sgroups.icmp_type_values_ok(value)
        );

drop type if exists sgroups.transport_entry cascade;
create type sgroups.transport_entry as (
  description text,
  comment     text,
  ports       sgroups.port_ranges,
  icmp_types  sgroups.icmp_types
);
comment on type sgroups.transport_entry is 'Transport entry: ports or ICMP types';

drop type if exists sgroups.transport cascade;
create type sgroups.transport as (
  proto   sgroups.proto,
  ip_v    sgroups.ip_family,
  entries sgroups.transport_entry[]
);
comment on type sgroups.transport is 'Transport: protocol, IP family, entries';

drop function if exists sgroups.transport_entries_ports_no_overlap(sgroups.transport[]) cascade;
create or replace function sgroups.transport_entries_ports_no_overlap(
  transports sgroups.transport[]
) returns boolean as $$
begin
  if transports is null then
    return true;
  end if;
  return not exists (
    select 1
    from unnest(transports) t
    where (t).proto in ('tcp', 'udp')
      and array_length((t).entries, 1) > 1
      and (
        select sgroups.i4mr_any_intersect((e).ports)
        from unnest((t).entries) e
      )
  );
end;
$$ language plpgsql immutable;
comment on function sgroups.transport_entries_ports_no_overlap
  is 'check that port ranges across entries within each transport do not overlap';

drop function if exists sgroups.transport_entries_icmp_no_overlap(sgroups.transport[]) cascade;
create or replace function sgroups.transport_entries_icmp_no_overlap(
  transports sgroups.transport[]
) returns boolean as $$
begin
  if transports is null then
    return true;
  end if;
  return not exists (
    select 1
    from unnest(transports) t
    where (t).proto = 'icmp'
      and array_length((t).entries, 1) > 1
      and exists (
        select 1
        from unnest((t).entries) as e1,
             unnest((t).entries) as e2
        where e1 is distinct from e2
          and (e1).icmp_types && (e2).icmp_types
      )
  );
end;
$$ language plpgsql immutable;
comment on function sgroups.transport_entries_icmp_no_overlap
  is 'check that ICMP types across entries within each transport do not overlap';

drop function if exists sgroups.check_service_transport_overlap(bigint, bigint[], sgroups.transport[]) cascade;
create or replace function sgroups.check_service_transport_overlap(
  svc_id bigint,
  ag_ids bigint[],
  new_transports sgroups.transport[]
)
returns boolean
as $$
begin
  if new_transports is null or array_length(new_transports, 1) is null then
    return false;
  end if;
  if ag_ids is null or array_length(ag_ids, 1) is null then
    return false;
  end if;

  return exists (
    with new_flat as (
      select (nt).proto, (nt).ip_v, (ne).ports, (ne).icmp_types
      from unnest(new_transports) as nt
      cross join lateral unnest((nt).entries) as ne
    ),
    existing_flat as (
      select (et).proto, (et).ip_v, (ee).ports, (ee).icmp_types
      from sgroups.tbl_service_binding sb
      join sgroups.tbl_service s on s.id = sb.service
      cross join lateral unnest(s.transports) as et
      cross join lateral unnest((et).entries) as ee
      where sb.ag = any(ag_ids)
        and s.id is distinct from svc_id
        and s.transports is not null
    )
    select 1
    from new_flat nf
    join existing_flat ef
      on ef.proto = nf.proto
     and ef.ip_v is not distinct from nf.ip_v
    where
      case
        when nf.proto in ('tcp', 'udp') then
          nf.ports is not null
          and ef.ports is not null
          and nf.ports && ef.ports
        else
          nf.icmp_types is not null
          and ef.icmp_types is not null
          and nf.icmp_types && ef.icmp_types
      end
    limit 1
  );
end;
$$ language plpgsql stable;

---------------------------------- RESOURCE TABLES -------------------------------------

drop table if exists sgroups.tbl_sync_status;
create table sgroups.tbl_sync_status(
    id bigint generated always as identity primary key,
    sync_count bigint not null,
    updated_at timestamptz generated always as ( sgroups.ts() ) stored,
    constraint sync_count_positive
         check (
            sync_count > 0
         )
);

insert into sgroups.tbl_sync_status(sync_count) (
    select 1 where not exists(select 1 from sgroups.tbl_sync_status)
);

drop table if exists sgroups.tbl_namespace cascade;
create table if not exists sgroups.tbl_namespace (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    display_name sgroups.dname default null,
    comment text default null,
    description text default null,
    labels hstore default null,
    annotations hstore default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,

    constraint ns_uid_uq  unique (uid),
    constraint ns_name_uq unique (name)
);

create index if not exists ns_labels_gin_idx
  on sgroups.tbl_namespace using gin (labels);

create index if not exists ns_annotations_gin_idx
  on sgroups.tbl_namespace using gin (annotations);

drop trigger if exists trg_namespace_immutable_fields on sgroups.tbl_namespace;
create trigger trg_namespace_immutable_fields
before update on sgroups.tbl_namespace
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_namespace_resource_version on sgroups.tbl_namespace;
create trigger trg_namespace_resource_version
before insert or update on sgroups.tbl_namespace
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_namespace_outbox_au on sgroups.tbl_namespace;
create trigger trg_namespace_outbox_au
after insert or update on sgroups.tbl_namespace
for each row
execute function sgroups.capture_outbox_resource_events_trg('Namespace', 'sgroups.vu_namespace');

drop trigger if exists trg_namespace_outbox_bd on sgroups.tbl_namespace;
create trigger trg_namespace_outbox_bd
before delete on sgroups.tbl_namespace
for each row
execute function sgroups.capture_outbox_resource_events_trg('Namespace', 'sgroups.vu_namespace');


drop table if exists sgroups.tbl_ag cascade;
create table if not exists sgroups.tbl_ag (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    default_action sgroups.policy_action not null default 'DENY',
    logs bool not null default false,
    trace bool not null default false,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    binding_rev bigint not null default 0,
    rule_rev bigint not null default 0,
    constraint ag_uid_uq  unique (uid),
    constraint ag_name_uq unique (name, ns),
    constraint fk_ag___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists ag_labels_gin_idx
  on sgroups.tbl_ag using gin (labels);

create index if not exists ag_annotations_gin_idx
  on sgroups.tbl_ag using gin (annotations);

drop trigger if exists trg_ag_immutable_fields on sgroups.tbl_ag;
create trigger trg_ag_immutable_fields
before update on sgroups.tbl_ag
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_ag_resource_version on sgroups.tbl_ag;
create trigger trg_ag_resource_version
before insert or update on sgroups.tbl_ag
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_ag_outbox_au on sgroups.tbl_ag;
create trigger trg_ag_outbox_au
after insert or update on sgroups.tbl_ag
for each row
execute function sgroups.capture_outbox_resource_events_trg('AddressGroup', 'sgroups.vu_ag');

drop trigger if exists trg_ag_outbox_bd on sgroups.tbl_ag;
create trigger trg_ag_outbox_bd
before delete on sgroups.tbl_ag
for each row
execute function sgroups.capture_outbox_resource_events_trg('AddressGroup', 'sgroups.vu_ag');


drop table if exists sgroups.tbl_network cascade;
create table sgroups.tbl_network (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    display_name sgroups.dname default null,
    network cidr not null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    binding_rev bigint not null default 0,
    constraint network_uid_uq unique (uid),
    constraint network_name_uniqueness
        unique (name, ns),
    constraint fk_network___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists nw_labels_gin_idx
  on sgroups.tbl_network using gin (labels);

create index if not exists nw_annotations_gin_idx
  on sgroups.tbl_network using gin (annotations);

drop trigger if exists trg_network_immutable_fields on sgroups.tbl_network;
create trigger trg_network_immutable_fields
before update on sgroups.tbl_network
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_network_resource_version on sgroups.tbl_network;
create trigger trg_network_resource_version
before insert or update on sgroups.tbl_network
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_network_outbox_au on sgroups.tbl_network;
create trigger trg_network_outbox_au
after insert or update on sgroups.tbl_network
for each row
execute function sgroups.capture_outbox_resource_events_trg('Network', 'sgroups.vu_network');

drop trigger if exists trg_network_outbox_bd on sgroups.tbl_network;
create trigger trg_network_outbox_bd
before delete on sgroups.tbl_network
for each row
execute function sgroups.capture_outbox_resource_events_trg('Network', 'sgroups.vu_network');



drop type if exists sgroups.host_info cascade;
create type sgroups.host_info as (
  host_name        sgroups.dname,
  os               sgroups.dname,
  platform         sgroups.dname,
  platform_family  sgroups.dname,
  platform_version sgroups.dname,
  kernel_version   sgroups.dname
);
comment on type sgroups.host_info
  is 'Host meta info';

drop table if exists sgroups.tbl_host cascade;
create table sgroups.tbl_host (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    display_name sgroups.dname default null,
    ips inet[] not null default '{}'::inet[],
    meta_info sgroups.host_info default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    binding_rev bigint not null default 0,
    constraint host_uid_uq  unique (uid),
    constraint host_name_uq unique (name, ns),
    constraint fk_host___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists host_labels_gin_idx
  on sgroups.tbl_host using gin (labels);

create index if not exists host_annotations_gin_idx
  on sgroups.tbl_host using gin (annotations);

drop trigger if exists trg_host_immutable_fields on sgroups.tbl_host;
create trigger trg_host_immutable_fields
before update on sgroups.tbl_host
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_host_resource_version on sgroups.tbl_host;
create trigger trg_host_resource_version
before insert or update on sgroups.tbl_host
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_host_outbox_au on sgroups.tbl_host;
create trigger trg_host_outbox_au
after insert or update on sgroups.tbl_host
for each row
execute function sgroups.capture_outbox_resource_events_trg('Host', 'sgroups.vu_host');

drop trigger if exists trg_host_outbox_bd on sgroups.tbl_host;
create trigger trg_host_outbox_bd
before delete on sgroups.tbl_host
for each row
execute function sgroups.capture_outbox_resource_events_trg('Host', 'sgroups.vu_host');



drop table if exists sgroups.tbl_host_binding cascade;
create table sgroups.tbl_host_binding (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    ag bigint,
    host bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    constraint host_binding_ag_host_uq unique (ag, host),
    constraint host_binding_uid_uq  unique (uid),
    constraint host_binding_name_uq unique (name, ns),
    constraint fk_host_binding___ag
       foreign key(ag) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_host_binding___host
       foreign key(host) references sgroups.tbl_host(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_host_binding___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists host_binding_labels_gin_idx
  on sgroups.tbl_host_binding using gin (labels);

create index if not exists host_binding_annotations_gin_idx
  on sgroups.tbl_host_binding using gin (annotations);

drop trigger if exists trg_host_binding_immutable_fields on sgroups.tbl_host_binding;
create trigger trg_host_binding_immutable_fields
before update on sgroups.tbl_host_binding
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name', 'ag', 'host');

drop trigger if exists trg_host_binding_resource_version on sgroups.tbl_host_binding;
create trigger trg_host_binding_resource_version
before insert or update on sgroups.tbl_host_binding
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_host_binding_outbox_au on sgroups.tbl_host_binding;
create trigger trg_host_binding_outbox_au
after insert or update on sgroups.tbl_host_binding
for each row
execute function sgroups.capture_outbox_resource_events_trg('HostBinding', 'sgroups.vu_host_binding');

drop trigger if exists trg_host_binding_outbox_bd on sgroups.tbl_host_binding;
create trigger trg_host_binding_outbox_bd
before delete on sgroups.tbl_host_binding
for each row
execute function sgroups.capture_outbox_resource_events_trg('HostBinding', 'sgroups.vu_host_binding');

drop trigger if exists trg_upd_hb_refs on sgroups.tbl_host_binding;
create trigger trg_upd_hb_refs
after update or insert or delete on sgroups.tbl_host_binding
for each row
execute function sgroups.upd_refs_trg(
  'binding_rev',
  'sgroups.tbl_host', 'host',
  'sgroups.tbl_ag',   'ag'
);


drop table if exists sgroups.tbl_network_binding cascade;
create table sgroups.tbl_network_binding (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    ag bigint,
    network bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    constraint network_binding_ag_network_uq unique (ag, network),
    constraint network_binding_uid_uq  unique (uid),
    constraint network_binding_name_uq unique (name, ns),
    constraint fk_network_binding___ag
       foreign key(ag) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_network_binding___network
       foreign key(network) references sgroups.tbl_network(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_network_binding___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists network_binding_labels_gin_idx
  on sgroups.tbl_network_binding using gin (labels);

create index if not exists network_binding_annotations_gin_idx
  on sgroups.tbl_network_binding using gin (annotations);

drop trigger if exists trg_network_binding_immutable_fields on sgroups.tbl_network_binding;
create trigger trg_network_binding_immutable_fields
before update on sgroups.tbl_network_binding
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name', 'ag', 'network');

drop trigger if exists trg_network_binding_resource_version on sgroups.tbl_network_binding;
create trigger trg_network_binding_resource_version
before insert or update on sgroups.tbl_network_binding
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_network_binding_outbox_au on sgroups.tbl_network_binding;
create trigger trg_network_binding_outbox_au
after insert or update on sgroups.tbl_network_binding
for each row
execute function sgroups.capture_outbox_resource_events_trg('NetworkBinding', 'sgroups.vu_network_binding');

drop trigger if exists trg_network_binding_outbox_bd on sgroups.tbl_network_binding;
create trigger trg_network_binding_outbox_bd
before delete on sgroups.tbl_network_binding
for each row
execute function sgroups.capture_outbox_resource_events_trg('NetworkBinding', 'sgroups.vu_network_binding');

drop trigger if exists trg_upd_nb_refs on sgroups.tbl_network_binding;
create trigger trg_upd_nb_refs
after update or insert or delete on sgroups.tbl_network_binding
for each row
execute function sgroups.upd_refs_trg(
  'binding_rev',
  'sgroups.tbl_network', 'network',
  'sgroups.tbl_ag',      'ag'
);


drop table if exists sgroups.tbl_service cascade;
create table sgroups.tbl_service (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    display_name sgroups.dname default null,
    transports sgroups.transport[] default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    binding_rev bigint not null default 0,
    rule_rev bigint not null default 0,
    constraint service_uid_uq unique (uid),
    constraint service_name_uq unique (name, ns),
    constraint "entries_ports_no_overlap"
         check (sgroups.transport_entries_ports_no_overlap(transports)),
    constraint "entries_icmp_no_overlap"
         check (sgroups.transport_entries_icmp_no_overlap(transports)),
    constraint fk_service___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists service_labels_gin_idx
  on sgroups.tbl_service using gin (labels);

create index if not exists service_annotations_gin_idx
  on sgroups.tbl_service using gin (annotations);

drop trigger if exists trg_service_immutable_fields on sgroups.tbl_service;
create trigger trg_service_immutable_fields
before update on sgroups.tbl_service
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name');

drop trigger if exists trg_service_resource_version on sgroups.tbl_service;
create trigger trg_service_resource_version
before insert or update on sgroups.tbl_service
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_service_outbox_au on sgroups.tbl_service;
create trigger trg_service_outbox_au
after insert or update on sgroups.tbl_service
for each row
execute function sgroups.capture_outbox_resource_events_trg('Service', 'sgroups.vu_service');

drop trigger if exists trg_service_outbox_bd on sgroups.tbl_service;
create trigger trg_service_outbox_bd
before delete on sgroups.tbl_service
for each row
execute function sgroups.capture_outbox_resource_events_trg('Service', 'sgroups.vu_service');

drop function if exists sgroups.check_transport_overlap_on_service_trg() cascade;
create or replace function sgroups.check_transport_overlap_on_service_trg()
returns trigger as $$
declare
  ag_ids bigint[];
begin
  if NEW.transports is not distinct from OLD.transports then
    return NEW;
  end if;
  select array_agg(sb.ag) into ag_ids
    from sgroups.tbl_service_binding sb where sb.service = NEW.id;

  if ag_ids is not null and
     sgroups.check_service_transport_overlap(NEW.id, ag_ids, NEW.transports) then
    raise exception 'service transport overlap on service "%"', NEW.name
      using detail = 'SG0010',
            hint = 'updated transports conflict with existing service bindings';
  end if;
  return NEW;
end;
$$ language plpgsql;

drop trigger if exists trg_service_transport_overlap on sgroups.tbl_service;
create trigger trg_service_transport_overlap
before update of transports on sgroups.tbl_service
for each row
execute function sgroups.check_transport_overlap_on_service_trg();


drop table if exists sgroups.tbl_service_binding cascade;
create table sgroups.tbl_service_binding (
    id bigint generated always as identity primary key,
    uid uuid not null default gen_random_uuid(),
    name sgroups.rname not null,
    ns bigint,
    ag bigint,
    service bigint,
    display_name sgroups.dname default null,
    labels hstore default null,
    annotations hstore default null,
    comment text default null,
    description text default null,
    creation_timestamp timestamptz not null default now(),
    resource_version text not null,
    constraint service_binding_ag_service_uq unique (ag, service),
    constraint service_binding_uid_uq  unique (uid),
    constraint service_binding_name_uq unique (name, ns),
    constraint fk_service_binding___ag
       foreign key(ag) references sgroups.tbl_ag(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_service_binding___service
       foreign key(service) references sgroups.tbl_service(id)
            on delete cascade
            on update restrict
            deferrable initially deferred,
    constraint fk_service_binding___ns
       foreign key(ns) references sgroups.tbl_namespace(id)
            on delete cascade
            on update restrict
            deferrable initially deferred
);

create index if not exists service_binding_labels_gin_idx
  on sgroups.tbl_service_binding using gin (labels);

create index if not exists service_binding_annotations_gin_idx
  on sgroups.tbl_service_binding using gin (annotations);

drop trigger if exists trg_service_binding_immutable_fields on sgroups.tbl_service_binding;
create trigger trg_service_binding_immutable_fields
before update on sgroups.tbl_service_binding
for each row
execute function sgroups.forbid_fields_update_trg('uid', 'name', 'ag', 'service');

drop trigger if exists trg_service_binding_resource_version on sgroups.tbl_service_binding;
create trigger trg_service_binding_resource_version
before insert or update on sgroups.tbl_service_binding
for each row
execute function sgroups.inc_resource_version_cnt();

drop trigger if exists trg_service_binding_outbox_au on sgroups.tbl_service_binding;
create trigger trg_service_binding_outbox_au
after insert or update on sgroups.tbl_service_binding
for each row
execute function sgroups.capture_outbox_resource_events_trg('ServiceBinding', 'sgroups.vu_service_binding');

drop trigger if exists trg_service_binding_outbox_bd on sgroups.tbl_service_binding;
create trigger trg_service_binding_outbox_bd
before delete on sgroups.tbl_service_binding
for each row
execute function sgroups.capture_outbox_resource_events_trg('ServiceBinding', 'sgroups.vu_service_binding');

drop trigger if exists trg_upd_sb_refs on sgroups.tbl_service_binding;
create trigger trg_upd_sb_refs
after update or insert or delete on sgroups.tbl_service_binding
for each row
execute function sgroups.upd_refs_trg(
  'binding_rev',
  'sgroups.tbl_service', 'service',
  'sgroups.tbl_ag',      'ag'
);

drop function if exists sgroups.check_transport_overlap_on_binding_trg() cascade;
create or replace function sgroups.check_transport_overlap_on_binding_trg()
returns trigger as $$
declare
  svc_transports sgroups.transport[];
begin
  select s.transports into svc_transports
    from sgroups.tbl_service s where s.id = NEW.service;

  if sgroups.check_service_transport_overlap(NEW.service, array[NEW.ag], svc_transports) then
    raise exception 'service transport overlap on binding "%"', NEW.name
      using detail = 'SG0010',
            hint = 'cannot bind service to AG when port/type ranges overlap with another bound service';
  end if;
  return NEW;
end;
$$ language plpgsql;

drop trigger if exists trg_service_binding_transport_overlap on sgroups.tbl_service_binding;
create constraint trigger trg_service_binding_transport_overlap
after insert on sgroups.tbl_service_binding
for each row
execute function sgroups.check_transport_overlap_on_binding_trg();

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

---------------------------------- RESOURCE LISTERS -------------------------------------

drop function if exists sgroups.list_namespaces(sgroups.res_selector[]) cascade;
create or replace function sgroups.list_namespaces(
  selectors sgroups.res_selector[] default null
)
returns setof sgroups.vu_namespace
as $$
begin
  return query
  select v.*
  from sgroups.vu_namespace v
  where sgroups.match_res_selectors(
    sgroups.mk_res_selector(
      v.name,
      null,
      null::sgroups.resource_ref[],
      v.labels
    ),
    selectors
  );
end;
$$ language plpgsql stable;


drop function if exists sgroups.list_ag(sgroups.res_selector[]) cascade;
create or replace function sgroups.list_ag(
  selectors sgroups.res_selector[] default null
)
returns setof sgroups.vu_ag
as $$
begin
  return query
  select v.*
  from sgroups.vu_ag v
  where sgroups.match_res_selectors(
    sgroups.mk_res_selector(
      v.name,
      v.namespace,
      v.refs,
      v.labels
    ),
    selectors
  );
end;
$$ language plpgsql stable;



drop function if exists sgroups.list_networks(sgroups.res_selector[]) cascade;
create or replace function sgroups.list_networks(
  selectors sgroups.res_selector[] default null
)
returns setof sgroups.vu_network
as $$
begin
  return query
  select v.*
  from sgroups.vu_network v
  where sgroups.match_res_selectors(
    sgroups.mk_res_selector(
      v.name,
      v.namespace,
      v.refs,
      v.labels
    ),
    selectors
  );
end;
$$ language plpgsql stable;


drop function if exists sgroups.list_hosts(sgroups.res_selector[]) cascade;
create or replace function sgroups.list_hosts(
  selectors sgroups.res_selector[] default null
)
returns setof sgroups.vu_host
as $$
begin
  return query
  select v.*
  from sgroups.vu_host v
  where sgroups.match_res_selectors(
    sgroups.mk_res_selector(
      v.name,
      v.namespace,
      v.refs,
      v.labels
    ),
    selectors
  );
end;
$$ language plpgsql stable;


drop function if exists sgroups.list_host_bindings() cascade;
create or replace function sgroups.list_host_bindings()
returns setof sgroups.vu_host_binding
as $$
begin
  return query
  select * from sgroups.vu_host_binding;
end;
$$ language plpgsql stable;


drop function if exists sgroups.list_network_bindings() cascade;
create or replace function sgroups.list_network_bindings()
returns setof sgroups.vu_network_binding
as $$
begin
  return query
  select * from sgroups.vu_network_binding;
end;
$$ language plpgsql stable;


drop function if exists sgroups.list_services(sgroups.res_selector[]) cascade;
create or replace function sgroups.list_services(
  selectors sgroups.res_selector[] default null
)
returns setof sgroups.vu_service
as $$
begin
  return query
  select v.*
  from sgroups.vu_service v
  where sgroups.match_res_selectors(
    sgroups.mk_res_selector(
      v.name,
      v.namespace,
      v.refs,
      v.labels
    ),
    selectors
  );
end;
$$ language plpgsql stable;


drop function if exists sgroups.list_service_bindings() cascade;
create or replace function sgroups.list_service_bindings()
returns setof sgroups.vu_service_binding
as $$
begin
  return query
  select * from sgroups.vu_service_binding;
end;
$$ language plpgsql stable;

---------------------------------- RESOURCE SYNCERS -------------------------------------

drop type if exists sgroups.row_of__namespace cascade;
create type sgroups.row_of__namespace as (
  uid uuid,
  name text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname
);

drop function if exists sgroups.sync_namespaces(sgroups.sync_op, sgroups.row_of__namespace) cascade;
create or replace function sgroups.sync_namespaces(
  op sgroups.sync_op, d sgroups.row_of__namespace
)
returns setof sgroups.vu_namespace
as $$
declare
  nsID bigint;
  affected_uid uuid;
  updated_count integer;
  norm_name text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');

  if op = 'del' then
    delete from sgroups.tbl_namespace t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid) and norm_name is not null and t.name = norm_name::sgroups.rname
    )
    returning id, uid into nsID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'namespace not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete you must pass existing uid OR matching name';
      end if;
      if norm_name is not null then
        raise exception 'namespace not found for name=%', norm_name
          using detail = 'SG0009', hint = 'for delete you must pass existing uid OR matching name';
      end if;
      raise exception 'namespace delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete you must pass existing uid OR matching name';
    end if;
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_namespaces: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        insert into sgroups.tbl_namespace (
          uid, name, display_name, comment, description, labels, annotations
        ) values (
          gen_random_uuid(),
          norm_name::sgroups.rname,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations
        )
        returning id, uid into nsID, affected_uid;
      exception
        when unique_violation then
          raise exception 'namespace already exists for name=%', norm_name
            using detail = 'SG0008', hint = 'for insert pass unique name; for update pass existing uid AND matching name';
      end;

      return query
      select v.*
      from sgroups.vu_namespace v
      where v.uid = affected_uid;
      return;
    end if;

    update sgroups.tbl_namespace t
       set display_name = (d).display_name,
           comment = (d).comment,
           description = (d).description,
           labels = (d).labels,
           annotations = (d).annotations
     where t.uid = (d).uid
       and t.name = (d).name::sgroups.rname;

    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_namespaces: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_namespace v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_namespaces: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__ag cascade;
create type sgroups.row_of__ag as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  default_action sgroups.policy_action,
  logs bool,
  trace bool
);


drop function if exists sgroups.sync_address_groups(sgroups.sync_op, sgroups.row_of__ag) cascade;
create or replace function sgroups.sync_address_groups(
  op sgroups.sync_op, d sgroups.row_of__ag
)
returns setof sgroups.vu_ag
as $$
declare
  affected_uid uuid;
  agID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');

  if op = 'del' then
    norm_ns := nullif(btrim((d).namespace), '');
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_address_groups: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    if sgroups.is_empty_uuid((d).uid) and norm_name is not null and norm_ns is not null then
      select id
        into nsID
        from sgroups.tbl_namespace
       where name = norm_ns::sgroups.rname;
      if not found then
        raise exception 'namespace not found for name=%', norm_ns
          using detail = 'SG0009', hint = 'pass existing namespace name';
      end if;
    end if;

    delete from sgroups.tbl_ag t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into agID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'address_group not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'address_group not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'address_group delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_address_groups: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    norm_ns := nullif(btrim((d).namespace), '');
    if norm_ns is null then
      raise exception 'sgroups.sync_address_groups: namespace is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing namespace for both insert (uid=null) and update (uid!=null)';
    end if;

    select id
      into nsID
      from sgroups.tbl_namespace
     where name = norm_ns::sgroups.rname;
    if not found then
      raise exception 'namespace not found for name=%', norm_ns
        using detail = 'SG0009', hint = 'pass existing namespace name';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        insert into sgroups.tbl_ag(
          uid, name, ns, display_name, comment, description, labels, annotations, default_action, logs, trace
        ) values (
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations,
          (d).default_action,
          (d).logs,
          (d).trace
        )
        returning id, uid into agID, affected_uid;
      exception
        when unique_violation then
          raise exception 'address_group already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_ag v
      where v.uid = affected_uid;
      return;
    end if;

    update sgroups.tbl_ag t
       set display_name = (d).display_name,
           comment = (d).comment,
           description = (d).description,
           labels = (d).labels,
           annotations = (d).annotations,
           default_action = (d).default_action,
           logs = (d).logs,
           trace = (d).trace
     where t.uid = (d).uid
       and t.name = (d).name::sgroups.rname
       and t.ns = nsID;

    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_address_groups: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_ag v
    where v.uid = (d).uid;
    return;
  end if;

raise exception 'sgroups.sync_address_groups: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__network cascade;
create type sgroups.row_of__network as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  network cidr
);


drop function if exists sgroups.sync_networks(sgroups.sync_op, sgroups.row_of__network) cascade;
create or replace function sgroups.sync_networks(
  op sgroups.sync_op, d sgroups.row_of__network
)
returns setof sgroups.vu_network
as $$
declare
  affected_uid uuid;
  nwID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');

  if op = 'del' then
    norm_ns := nullif(btrim((d).namespace), '');
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_networks: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    if sgroups.is_empty_uuid((d).uid) and norm_name is not null and norm_ns is not null then
      select id
        into nsID
        from sgroups.tbl_namespace
       where name = norm_ns::sgroups.rname;
      if not found then
        raise exception 'namespace not found for name=%', norm_ns
          using detail = 'SG0009', hint = 'pass existing namespace name';
      end if;
    end if;

    delete from sgroups.tbl_network t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into nwID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'network not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'network not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'network delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_networks: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    norm_ns := nullif(btrim((d).namespace), '');
    if norm_ns is null then
      raise exception 'sgroups.sync_networks: namespace is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing namespace for both insert (uid=null) and update (uid!=null)';
    end if;

    if (d).network is null then
      raise exception 'sgroups.sync_networks: network is required for op=ups'
        using detail = 'SG0007', hint = 'pass CIDR value for both insert and update';
    end if;

    select id
      into nsID
      from sgroups.tbl_namespace
     where name = norm_ns::sgroups.rname;
    if not found then
      raise exception 'namespace not found for name=%', norm_ns
        using detail = 'SG0009', hint = 'pass existing namespace name';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        insert into sgroups.tbl_network(
          uid, name, ns, display_name, comment, description, labels, annotations, network
        ) values (
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations,
          (d).network
        )
        returning id, uid into nwID, affected_uid;
      exception
        when unique_violation then
          raise exception 'network already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_network v
      where v.uid = affected_uid;
      return;
    end if;

    update sgroups.tbl_network t
       set display_name = (d).display_name,
           comment = (d).comment,
           description = (d).description,
           labels = (d).labels,
           annotations = (d).annotations,
           network = (d).network
     where t.uid = (d).uid
       and t.name = (d).name::sgroups.rname
       and t.ns = nsID;

    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_networks: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_network v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_networks: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__host cascade;
create type sgroups.row_of__host as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname
);


drop function if exists sgroups.sync_hosts(sgroups.sync_op, sgroups.row_of__host) cascade;
create or replace function sgroups.sync_hosts(
  op sgroups.sync_op, d sgroups.row_of__host
)
returns setof sgroups.vu_host
as $$
declare
  affected_uid uuid;
  hostID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');

  if op = 'del' then
    norm_ns := nullif(btrim((d).namespace), '');
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_hosts: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    if sgroups.is_empty_uuid((d).uid) and norm_name is not null and norm_ns is not null then
      select id
        into nsID
        from sgroups.tbl_namespace
       where name = norm_ns::sgroups.rname;
      if not found then
        raise exception 'namespace not found for name=%', norm_ns
          using detail = 'SG0009', hint = 'pass existing namespace name';
      end if;
    end if;

    delete from sgroups.tbl_host t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into hostID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'host not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'host not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'host delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_hosts: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    norm_ns := nullif(btrim((d).namespace), '');
    if norm_ns is null then
      raise exception 'sgroups.sync_hosts: namespace is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing namespace for both insert (uid=null) and update (uid!=null)';
    end if;

    select id
      into nsID
      from sgroups.tbl_namespace
     where name = norm_ns::sgroups.rname;
    if not found then
      raise exception 'namespace not found for name=%', norm_ns
        using detail = 'SG0009', hint = 'pass existing namespace name';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        insert into sgroups.tbl_host(
          uid, name, ns, display_name, comment, description, labels, annotations
        ) values (
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations
        )
        returning id, uid into hostID, affected_uid;
      exception
        when unique_violation then
          raise exception 'host already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_host v
      where v.uid = affected_uid;
      return;
    end if;

    update sgroups.tbl_host t
       set display_name = (d).display_name,
           comment = (d).comment,
           description = (d).description,
           labels = (d).labels,
           annotations = (d).annotations
     where t.uid = (d).uid
       and t.name = (d).name::sgroups.rname
       and t.ns = nsID;

    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_hosts: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_host v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_hosts: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__host_ips cascade;
create type sgroups.row_of__host_ips as (
  uid uuid,
  name text,
  namespace text,
  ips  inet[]
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
     set ips = coalesce((d).ips, '{}'::inet[])
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


drop type if exists sgroups.row_of__host_info cascade;
create type sgroups.row_of__host_info as (
  uid uuid,
  name text,
  namespace text,
  meta_info  sgroups.host_info
);

drop function if exists sgroups.sync_host_info(sgroups.sync_op, sgroups.row_of__host_info) cascade;
create or replace function sgroups.sync_host_info(op sgroups.sync_op, d sgroups.row_of__host_info)
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
    raise exception 'sgroups.sync_host_info: unsupported op=% (use ups)', op;
  end if;

  norm_name := nullif(btrim((d).name), '');
  norm_ns := nullif(btrim((d).namespace), '');

  if sgroups.is_empty_uuid((d).uid) or norm_name is null or norm_ns is null then
      raise exception 'sgroups.sync_host_info: uid, name and namespace are required for updating host meta info'
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
   where not sgroups.is_empty_uuid((d).uid)
      and t.uid = (d).uid
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID;
  if hostID is null then
    raise exception 'host not found for name=% namespace=% and uid=%', norm_name, norm_ns, (d).uid
      using detail = 'SG0009', hint = 'pass existing name, namespace and uid';
  end if;

  update sgroups.tbl_host t
     set meta_info = (d).meta_info
   where t.id = hostID;

  get diagnostics updated_count = row_count;
  if updated_count <> 1 then
    raise exception 'sgroups.sync_host_info: update mismatch (updated=%)', updated_count
      using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name AND namespace';
  end if;


  return query
  select v.*
  from sgroups.vu_host v
  where v.uid = (d).uid;
  return;

end;
$$ language plpgsql strict;


drop function if exists sgroups.resolve_host_binding_targets(text, sgroups.resource_id, sgroups.resource_id) cascade;
create or replace function sgroups.resolve_host_binding_targets(
  hbNs text, ag sgroups.resource_id, host sgroups.resource_id
)
returns table(ag_id bigint, host_id bigint)
as $$
declare
  hostID bigint;
  hostName text;
  hostNs text;
  agID bigint;
  agName text;
  agNs text;
begin
    if ag is null then
      raise exception 'sgroups.resolve_host_binding_targets: address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing address group reference';
    end if;
    if host is null then
      raise exception 'sgroups.resolve_host_binding_targets: host reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing host reference';
    end if;

    agName := nullif(btrim((ag).name), '');
    agNs := nullif(btrim((ag).namespace), '');
    hostName := nullif(btrim((host).name), '');
    hostNs := nullif(btrim((host).namespace), '');

    if agName is null or agNs is null then
      raise exception 'sgroups.resolve_host_binding_targets: both address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing address group name and namespace';
    end if;
    if hostName is null or hostNs is null then
      raise exception 'sgroups.resolve_host_binding_targets: both host name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing host name and namespace';
    end if;

    if hostNs <> agNs then
      raise exception 'sgroups.resolve_host_binding_targets: host namespace and address group namespace must match'
        using detail = 'SG0007', hint = 'pass host and address group in the same namespace';
    end if;

    if hbNs <> hostNs then
      raise exception 'sgroups.resolve_host_binding_targets: host binding namespace must match host namespace'
        using detail = 'SG0007', hint = 'pass host binding namespace that matches host namespace';
    end if;

    select t.id
    from sgroups.tbl_host t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = hostName::sgroups.rname
      and ns.name = hostNs::sgroups.rname
    into hostID;
    if hostID is null then
      raise exception 'host not found for name=% namespace=%', hostName, hostNs
        using detail = 'SG0009', hint = 'pass existing host name and namespace';
    end if;

    select t.id
    from sgroups.tbl_ag t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = agName::sgroups.rname
      and ns.name = agNs::sgroups.rname
    into agID;
    if agID is null then
      raise exception 'address group not found for name=% namespace=%', agName, agNs
        using detail = 'SG0009', hint = 'pass existing address group name and namespace';
    end if;

    ag_id := agID;
    host_id := hostID;
    return next;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__host_binding cascade;
create type sgroups.row_of__host_binding as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  ag sgroups.resource_id,
  host sgroups.resource_id
);

drop function if exists sgroups.sync_host_bindings(sgroups.sync_op, sgroups.row_of__host_binding) cascade;
create or replace function sgroups.sync_host_bindings(
  op sgroups.sync_op, d sgroups.row_of__host_binding
)
returns setof sgroups.vu_host_binding
as $$
declare
  affected_uid uuid;
  hbID bigint;
  nsID bigint;
  updated_count integer;
  constraint_name text;
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
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_host_bindings: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    delete from sgroups.tbl_host_binding t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into hbID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'host binding not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'host binding not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'host binding delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_host_bindings: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_host_bindings: namespace is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing namespace for both insert (uid=null) and update (uid!=null)';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        with ids as (
          select * from sgroups.resolve_host_binding_targets(norm_ns, (d).ag, (d).host)
        )
        insert into sgroups.tbl_host_binding(
          uid, name, ns, display_name, comment, description, labels, annotations, ag, host
        )
        select
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations,
          ids.ag_id,
          ids.host_id
        from ids
        returning id, uid into hbID, affected_uid;
      exception
        when unique_violation then
          get stacked diagnostics constraint_name = CONSTRAINT_NAME;
          if constraint_name = 'host_binding_ag_host_uq' then
            raise exception 'host binding already exists for given (address_group, host)'
              using detail = 'SG0008', hint = 'binding (address_group, host) must be unique';
          end if;
          raise exception 'host binding already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_host_binding v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_host_binding_targets(norm_ns, (d).ag, (d).host)
      )
      update sgroups.tbl_host_binding t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             ag = ids.ag_id,
             host = ids.host_id
        from ids
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into hbID;
    exception
      when unique_violation then
        get stacked diagnostics constraint_name = CONSTRAINT_NAME;
        if constraint_name = 'host_binding_ag_host_uq' then
          raise exception 'host binding already exists for given (address_group, host)'
            using detail = 'SG0008', hint = 'binding (address_group, host) must be unique';
        end if;

        raise exception 'host binding already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_host_bindings: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_host_binding v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_host_bindings: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop function if exists sgroups.resolve_network_binding_targets(text, sgroups.resource_id, sgroups.resource_id) cascade;
create or replace function sgroups.resolve_network_binding_targets(
  nbNs text, ag sgroups.resource_id, network sgroups.resource_id
)
returns table(ag_id bigint, nw_id bigint)
as $$
declare
  nwID bigint;
  nwName text;
  nwNs text;
  agID bigint;
  agName text;
  agNs text;
begin
    if ag is null then
      raise exception 'sgroups.resolve_network_binding_targets: address group reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing address group reference';
    end if;
    if network is null then
      raise exception 'sgroups.resolve_network_binding_targets: network reference is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing network reference';
    end if;

    agName := nullif(btrim((ag).name), '');
    agNs := nullif(btrim((ag).namespace), '');
    nwName := nullif(btrim((network).name), '');
    nwNs := nullif(btrim((network).namespace), '');

    if agName is null or agNs is null then
      raise exception 'sgroups.resolve_network_binding_targets: both address group name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing address group name and namespace';
    end if;
    if nwName is null or nwNs is null then
      raise exception 'sgroups.resolve_network_binding_targets: both network name and namespace is required'
        using detail = 'SG0007', hint = 'pass existing network name and namespace';
    end if;

    if nwNs <> agNs then
      raise exception 'sgroups.resolve_network_binding_targets: network namespace and address group namespace must match'
        using detail = 'SG0007', hint = 'pass network and address group in the same namespace';
    end if;

    if nbNs <> nwNs then
      raise exception 'sgroups.resolve_network_binding_targets: network binding namespace must match network namespace'
        using detail = 'SG0007', hint = 'pass network binding namespace that matches network namespace';
    end if;

    select t.id
    from sgroups.tbl_network t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = nwName::sgroups.rname
      and ns.name = nwNs::sgroups.rname
    into nwID;
    if nwID is null then
      raise exception 'network not found for name=% namespace=%', nwName, nwNs
        using detail = 'SG0009', hint = 'pass existing network name and namespace';
    end if;

    select t.id
    from sgroups.tbl_ag t
    join sgroups.tbl_namespace ns on ns.id = t.ns
    where t.name = agName::sgroups.rname
      and ns.name = agNs::sgroups.rname
    into agID;
    if agID is null then
      raise exception 'address group not found for name=% namespace=%', agName, agNs
        using detail = 'SG0009', hint = 'pass existing address group name and namespace';
    end if;

    ag_id := agID;
    nw_id := nwID;
    return next;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__network_binding cascade;
create type sgroups.row_of__network_binding as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  ag sgroups.resource_id,
  network sgroups.resource_id
);

drop function if exists sgroups.sync_network_bindings(sgroups.sync_op, sgroups.row_of__network_binding) cascade;
create or replace function sgroups.sync_network_bindings(
  op sgroups.sync_op, d sgroups.row_of__network_binding
)
returns setof sgroups.vu_network_binding
as $$
declare
  affected_uid uuid;
  nbID bigint;
  nsID bigint;
  updated_count integer;
  constraint_name text;
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
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_network_bindings: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    delete from sgroups.tbl_network_binding t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into nbID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'network binding not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'network binding not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'network binding delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    if norm_name is null then
      raise exception 'sgroups.sync_network_bindings: name is required for op=ups'
        using detail = 'SG0007', hint = 'set name for both insert (uid=null) and update (uid!=null)';
    end if;

    if norm_ns is null then
      raise exception 'sgroups.sync_network_bindings: namespace is required for op=ups'
        using detail = 'SG0007', hint = 'pass existing namespace for both insert (uid=null) and update (uid!=null)';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        with ids as (
          select * from sgroups.resolve_network_binding_targets(norm_ns, (d).ag, (d).network)
        )
        insert into sgroups.tbl_network_binding(
          uid, name, ns, display_name, comment, description, labels, annotations, ag, network
        )
        select
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations,
          ids.ag_id,
          ids.nw_id
        from ids
        returning id, uid into nbID, affected_uid;
      exception
        when unique_violation then
          get stacked diagnostics constraint_name = CONSTRAINT_NAME;
          if constraint_name = 'network_binding_ag_network_uq' then
            raise exception 'network binding already exists for given (address_group, network)'
              using detail = 'SG0008', hint = 'binding (address_group, network) must be unique';
          end if;
          raise exception 'network binding already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
      end;

      return query
      select v.*
      from sgroups.vu_network_binding v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_network_binding_targets(norm_ns, (d).ag, (d).network)
      )
      update sgroups.tbl_network_binding t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             ag = ids.ag_id,
             network = ids.nw_id
        from ids
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID
       returning id into nbID;
    exception
      when unique_violation then
        get stacked diagnostics constraint_name = CONSTRAINT_NAME;
        if constraint_name = 'network_binding_ag_network_uq' then
          raise exception 'network binding already exists for given (address_group, network)'
            using detail = 'SG0008', hint = 'binding (address_group, network) must be unique';
        end if;

        raise exception 'network binding already exists for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0008',
                hint   = 'for insert pass unique (name, namespace); for update pass existing uid AND matching name AND namespace';
    end;
    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_network_bindings: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update you must pass existing uid AND matching name';
    end if;

    return query
    select v.*
    from sgroups.vu_network_binding v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_network_bindings: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop function if exists sgroups.resolve_service_binding_targets(
  text, sgroups.resource_id, sgroups.resource_id
) cascade;
create or replace function sgroups.resolve_service_binding_targets(
  sbNs text,
  ag sgroups.resource_id,
  service sgroups.resource_id
)
returns table (ag_id bigint, service_id bigint)
as $$
declare
  agNs text;
  svcNs text;
  resolved_ag_id bigint;
  resolved_svc_id bigint;
begin
  agNs := nullif(btrim((ag).namespace), '');
  svcNs := nullif(btrim((service).namespace), '');

  -- Service must be in the same namespace as the binding; AG can be in any namespace
  if svcNs is distinct from sbNs then
    raise exception 'service_binding: service must be in the same namespace as binding (binding_ns=%, service_ns=%)',
      sbNs, svcNs
      using detail = 'SG0007',
            hint   = 'binding and service must be in the same namespace; address_group can be in a different namespace';
  end if;

  select a.id into resolved_ag_id
    from sgroups.tbl_ag a
    join sgroups.tbl_namespace ns on ns.id = a.ns
   where a.name = (ag).name::sgroups.rname
     and ns.name = agNs::sgroups.rname;

  if not found then
    raise exception 'address_group not found: name=% namespace=%', (ag).name, agNs
      using detail = 'SG0009', hint = 'pass existing address_group name and namespace';
  end if;

  select s.id into resolved_svc_id
    from sgroups.tbl_service s
    join sgroups.tbl_namespace ns on ns.id = s.ns
   where s.name = (service).name::sgroups.rname
     and ns.name = svcNs::sgroups.rname;

  if not found then
    raise exception 'service not found: name=% namespace=%', (service).name, svcNs
      using detail = 'SG0009', hint = 'pass existing service name and namespace';
  end if;

  return query select resolved_ag_id, resolved_svc_id;
end;
$$ language plpgsql stable;


drop type if exists sgroups.row_of__service cascade;
create type sgroups.row_of__service as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  transports sgroups.transport[]
);

drop function if exists sgroups.sync_services(sgroups.sync_op, sgroups.row_of__service) cascade;
create or replace function sgroups.sync_services(
  op sgroups.sync_op, d sgroups.row_of__service
)
returns setof sgroups.vu_service
as $$
declare
  affected_uid uuid;
  svcID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');

  if op = 'del' then
    norm_ns := nullif(btrim((d).namespace), '');
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_services: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    if sgroups.is_empty_uuid((d).uid) and norm_name is not null and norm_ns is not null then
      select id
        into nsID
        from sgroups.tbl_namespace
       where name = norm_ns::sgroups.rname;
      if not found then
        raise exception 'namespace not found for name=%', norm_ns
          using detail = 'SG0009', hint = 'pass existing namespace name';
      end if;
    end if;

    delete from sgroups.tbl_service t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into svcID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'service not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'service not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'service delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    norm_ns := nullif(btrim((d).namespace), '');
    if norm_name is null or norm_ns is null then
      raise exception 'sgroups.sync_services: name and namespace are required for op=ups'
        using detail = 'SG0007', hint = 'set name and namespace';
    end if;

    select id into nsID
      from sgroups.tbl_namespace
     where name = norm_ns::sgroups.rname;

    if not found then
      raise exception 'namespace not found for name=%', norm_ns
        using detail = 'SG0009', hint = 'pass existing namespace name';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      begin
        insert into sgroups.tbl_service (
          uid, name, ns, display_name, comment, description, labels, annotations, transports
        ) values (
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations,
          (d).transports
        )
        returning id, uid into svcID, affected_uid;
      exception
        when unique_violation then
          raise exception 'service already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique name; for update pass existing uid AND matching name';
      end;

      return query
      select v.*
      from sgroups.vu_service v
      where v.uid = affected_uid;
      return;
    end if;

    update sgroups.tbl_service t
       set display_name = (d).display_name,
           comment = (d).comment,
           description = (d).description,
           labels = (d).labels,
           annotations = (d).annotations,
           transports = (d).transports
     where t.uid = (d).uid
       and t.name = (d).name::sgroups.rname
       and t.ns = nsID;

    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_services: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update pass existing uid AND matching name/namespace';
    end if;

    return query
    select v.*
    from sgroups.vu_service v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_services: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;


drop type if exists sgroups.row_of__service_binding cascade;
create type sgroups.row_of__service_binding as (
  uid uuid,
  name text,
  namespace text,
  labels hstore,
  annotations hstore,
  comment text,
  description text,
  display_name sgroups.dname,
  ag sgroups.resource_id,
  service sgroups.resource_id
);

drop function if exists sgroups.sync_service_bindings(sgroups.sync_op, sgroups.row_of__service_binding) cascade;
create or replace function sgroups.sync_service_bindings(
  op sgroups.sync_op, d sgroups.row_of__service_binding
)
returns setof sgroups.vu_service_binding
as $$
declare
  affected_uid uuid;
  sbID bigint;
  nsID bigint;
  updated_count integer;
  norm_name text;
  norm_ns text;
begin
  if d is null then
    return;
  end if;

  norm_name := nullif(btrim((d).name), '');

  if op = 'del' then
    norm_ns := nullif(btrim((d).namespace), '');
    if sgroups.is_empty_uuid((d).uid) and (norm_name is null or norm_ns is null) then
      raise exception 'sgroups.sync_service_bindings: name and namespace are required for op=del when uid is empty'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;

    if sgroups.is_empty_uuid((d).uid) and norm_name is not null and norm_ns is not null then
      select id
        into nsID
        from sgroups.tbl_namespace
       where name = norm_ns::sgroups.rname;
      if not found then
        raise exception 'namespace not found for name=%', norm_ns
          using detail = 'SG0009', hint = 'pass existing namespace name';
      end if;
    end if;

    delete from sgroups.tbl_service_binding t
    where (
      not sgroups.is_empty_uuid((d).uid) and t.uid = (d).uid
    ) or (
      sgroups.is_empty_uuid((d).uid)
      and norm_name is not null
      and nsID is not null
      and t.name = norm_name::sgroups.rname
      and t.ns = nsID
    )
    returning id, uid into sbID, affected_uid;

    if not found then
      if not sgroups.is_empty_uuid((d).uid) then
        raise exception 'service_binding not found for uid=%', (d).uid
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      if norm_name is not null then
        raise exception 'service_binding not found for name=% namespace=%', norm_name, norm_ns
          using detail = 'SG0009', hint = 'for delete pass existing uid OR (name AND namespace)';
      end if;
      raise exception 'service_binding delete: either uid or name is required'
        using detail = 'SG0007', hint = 'for delete pass existing uid OR (name AND namespace)';
    end if;
    return;
  end if;

  if op = 'ups' then
    norm_ns := nullif(btrim((d).namespace), '');
    if norm_name is null or norm_ns is null then
      raise exception 'sgroups.sync_service_bindings: name and namespace are required for op=ups'
        using detail = 'SG0007', hint = 'set name and namespace';
    end if;

    select id into nsID
      from sgroups.tbl_namespace
     where name = norm_ns::sgroups.rname;

    if not found then
      raise exception 'namespace not found for name=%', norm_ns
        using detail = 'SG0009', hint = 'pass existing namespace name';
    end if;

    if sgroups.is_empty_uuid((d).uid) then
      declare
        resolved_ag_id bigint;
        resolved_svc_id bigint;
      begin
        select ag_id, service_id
          into resolved_ag_id, resolved_svc_id
          from sgroups.resolve_service_binding_targets(
            norm_ns, (d).ag, (d).service
          );

        insert into sgroups.tbl_service_binding (
          uid, name, ns, ag, service, display_name, comment, description, labels, annotations
        ) values (
          gen_random_uuid(),
          norm_name::sgroups.rname,
          nsID,
          resolved_ag_id,
          resolved_svc_id,
          (d).display_name,
          (d).comment,
          (d).description,
          (d).labels,
          (d).annotations
        )
        returning id, uid into sbID, affected_uid;
      exception
        when unique_violation then
          if sqlerrm ~* 'service_binding_ag_service_uq' then
            raise exception 'service_binding already exists for ag=(%, %) service=(%, %)',
              (d).ag.name, (d).ag.namespace, (d).service.name, (d).service.namespace
              using detail = 'SG0008', hint = 'duplicate binding: this (ag, service) pair already exists';
          end if;
          raise exception 'service_binding already exists for name=% namespace=%', norm_name, norm_ns
            using detail = 'SG0008', hint = 'for insert pass unique name; for update pass existing uid AND matching name';
      end;

      return query
      select v.*
      from sgroups.vu_service_binding v
      where v.uid = affected_uid;
      return;
    end if;

    begin
      with ids as (
        select * from sgroups.resolve_service_binding_targets(
          norm_ns, (d).ag, (d).service
        )
      )
      update sgroups.tbl_service_binding t
         set display_name = (d).display_name,
             comment = (d).comment,
             description = (d).description,
             labels = (d).labels,
             annotations = (d).annotations,
             ag = ids.ag_id,
             service = ids.service_id
        from ids
       where t.uid = (d).uid
         and t.name = (d).name::sgroups.rname
         and t.ns = nsID;
    exception
      when unique_violation then
        raise exception 'service_binding update conflict for uid=%', (d).uid
          using detail = 'SG0008', hint = 'for update pass existing uid AND matching name/namespace';
    end;

    get diagnostics updated_count = row_count;
    if updated_count <> 1 then
      raise exception 'sgroups.sync_service_bindings: update mismatch (updated=%)', updated_count
        using detail = 'SG0008', hint = 'for update pass existing uid AND matching name/namespace';
    end if;

    return query
    select v.*
    from sgroups.vu_service_binding v
    where v.uid = (d).uid;
    return;
  end if;

  raise exception 'sgroups.sync_service_bindings: unsupported op=% (use ups or del)', op;
end;
$$ language plpgsql strict;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop schema if exists sgroups cascade;

-- +goose StatementEnd
