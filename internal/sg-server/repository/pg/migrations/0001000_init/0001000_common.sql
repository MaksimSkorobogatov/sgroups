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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop schema if exists sgroups cascade;
-- +goose StatementEnd