-- +goose Up
-- +goose StatementBegin

-- display_name uniqueness: per-table for resources, cross-table (within ns) for rules.

drop view if exists sgroups.vu_rule_display_names cascade;
create or replace view sgroups.vu_rule_display_names as
          select uid, ns, display_name from sgroups.tbl_ag2ag_rule         where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2ag_icmp_rule    where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2icmp_rule       where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2cidr_rule       where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2cidr_icmp_rule  where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2fqdn_rule       where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_svc2svc_rule       where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_svc2fqdn_rule      where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_svc2cidr_rule      where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_svc2cidr_icmp_rule where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2svc_rule        where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_svc2ag_rule        where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_svc2ag_icmp_rule   where display_name is not null
union all select uid, ns, display_name from sgroups.tbl_ag2svc_icmp_rule   where display_name is not null;

do $$
declare
  t text;
  tables_all text[] := array[
    'sgroups.tbl_namespace',
    'sgroups.tbl_ag',
    'sgroups.tbl_network',
    'sgroups.tbl_host',
    'sgroups.tbl_host_binding',
    'sgroups.tbl_network_binding',
    'sgroups.tbl_service',
    'sgroups.tbl_service_binding',
    'sgroups.tbl_ag2ag_rule',
    'sgroups.tbl_ag2ag_icmp_rule',
    'sgroups.tbl_ag2icmp_rule',
    'sgroups.tbl_ag2cidr_rule',
    'sgroups.tbl_ag2cidr_icmp_rule',
    'sgroups.tbl_ag2fqdn_rule',
    'sgroups.tbl_svc2svc_rule',
    'sgroups.tbl_svc2fqdn_rule',
    'sgroups.tbl_svc2cidr_rule',
    'sgroups.tbl_svc2cidr_icmp_rule',
    'sgroups.tbl_ag2svc_rule',
    'sgroups.tbl_svc2ag_rule',
    'sgroups.tbl_svc2ag_icmp_rule',
    'sgroups.tbl_ag2svc_icmp_rule'
  ];
  tables_non_rule_ns text[] := array[
    'sgroups.tbl_ag',
    'sgroups.tbl_network',
    'sgroups.tbl_host',
    'sgroups.tbl_host_binding',
    'sgroups.tbl_network_binding',
    'sgroups.tbl_service',
    'sgroups.tbl_service_binding'
  ];
begin
  foreach t in array tables_all loop
    execute format($f$
      update %s
         set display_name = nullif(
           btrim(
             regexp_replace(
               regexp_replace(
                 regexp_replace(lower(display_name), '[^a-z0-9-]', '-', 'g'),
                 '-+', '-', 'g'),
               '(^-+)|(-+$)', '', 'g'),
             '-'),
           '')::sgroups.dname
       where display_name is not null
         and display_name !~ '^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$';
    $f$, t);

    execute format($f$
      update %s
         set display_name = nullif(substring(display_name from 1 for 63), '')::sgroups.dname
       where display_name is not null and length(display_name) > 63;
    $f$, t);
  end loop;

  with d as (
    select id, display_name,
           row_number() over (partition by display_name order by id) as rn
      from sgroups.tbl_namespace
     where display_name is not null
  )
  update sgroups.tbl_namespace t
     set display_name = (
       substring(d.display_name from 1 for 63 - length('-' || d.id::text))
       || '-' || d.id::text
     )::sgroups.dname
    from d
   where t.id = d.id and d.rn > 1;

  foreach t in array tables_non_rule_ns loop
    execute format($f$
      with d as (
        select id, display_name,
               row_number() over (partition by ns, display_name order by id) as rn
          from %s
         where display_name is not null
      )
      update %s t
         set display_name = (
           substring(d.display_name from 1 for 63 - length('-' || d.id::text))
           || '-' || d.id::text
         )::sgroups.dname
        from d
       where t.id = d.id and d.rn > 1;
    $f$, t, t);
  end loop;
end$$;

do $$
declare
  t text;
  tables_rule text[] := array[
    'sgroups.tbl_ag2ag_rule',        'sgroups.tbl_ag2ag_icmp_rule',
    'sgroups.tbl_ag2icmp_rule',      'sgroups.tbl_ag2cidr_rule',
    'sgroups.tbl_ag2cidr_icmp_rule', 'sgroups.tbl_ag2fqdn_rule',
    'sgroups.tbl_svc2svc_rule',      'sgroups.tbl_svc2fqdn_rule',
    'sgroups.tbl_svc2cidr_rule',     'sgroups.tbl_svc2cidr_icmp_rule',
    'sgroups.tbl_ag2svc_rule',       'sgroups.tbl_svc2ag_rule',
    'sgroups.tbl_svc2ag_icmp_rule',  'sgroups.tbl_ag2svc_icmp_rule'
  ];
begin
  create temp table _rule_dn_dedup on commit drop as
  with ranked as (
    select uid, ns, display_name,
           row_number() over (partition by ns, display_name order by uid) as rn
      from sgroups.vu_rule_display_names
  )
  select uid,
         (
           substring(display_name::text from 1 for 63 - length('-' || replace(uid::text, '-', '')))
           || '-' || replace(uid::text, '-', '')
         )::sgroups.dname as new_display_name
    from ranked
   where rn > 1;

  foreach t in array tables_rule loop
    execute format(
      'update %s tgt set display_name = d.new_display_name '
      'from _rule_dn_dedup d where tgt.uid = d.uid', t);
  end loop;
end$$;

-- flush deferred FK events from the dedupe updates so ADD CONSTRAINT can run
set constraints all immediate;

alter table sgroups.tbl_namespace
  add constraint ns_display_name_uq unique (display_name);

alter table sgroups.tbl_ag
  add constraint ag_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_network
  add constraint network_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_host
  add constraint host_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_host_binding
  add constraint host_binding_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_network_binding
  add constraint network_binding_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_service
  add constraint service_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_service_binding
  add constraint service_binding_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2ag_rule
  add constraint ag2ag_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2ag_icmp_rule
  add constraint ag2ag_icmp_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2icmp_rule
  add constraint ag2icmp_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2cidr_rule
  add constraint ag2cidr_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2cidr_icmp_rule
  add constraint ag2cidr_icmp_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2fqdn_rule
  add constraint ag2fqdn_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_svc2svc_rule
  add constraint svc2svc_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_svc2fqdn_rule
  add constraint svc2fqdn_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_svc2cidr_rule
  add constraint svc2cidr_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_svc2cidr_icmp_rule
  add constraint svc2cidr_icmp_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2svc_rule
  add constraint ag2svc_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_svc2ag_rule
  add constraint svc2ag_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_svc2ag_icmp_rule
  add constraint svc2ag_icmp_rule_display_name_uq unique (display_name, ns);

alter table sgroups.tbl_ag2svc_icmp_rule
  add constraint ag2svc_icmp_rule_display_name_uq unique (display_name, ns);

drop function if exists sgroups.check_display_name_uniqueness_trg() cascade;
create or replace function sgroups.check_display_name_uniqueness_trg()
returns trigger
as $$
declare
  newj            jsonb;
  oldj            jsonb;
  dname           text;
  ns_val          bigint;
  conflict_exists boolean;
begin
  NEW.display_name := nullif(btrim(NEW.display_name::text), '')::sgroups.dname;
  if NEW.display_name is null then
    return NEW;
  end if;

  newj := to_jsonb(NEW);

  if TG_OP = 'UPDATE' then
    oldj := to_jsonb(OLD);
    if (newj ->> 'display_name') is not distinct from (oldj ->> 'display_name')
       and (newj ->> 'ns') is not distinct from (oldj ->> 'ns') then
      return NEW;
    end if;
  end if;

  dname := NEW.display_name::text;

  if newj ? 'ns' then
    ns_val := nullif(newj ->> 'ns', '')::bigint;
    if ns_val is null then
      return NEW;
    end if;
    execute format(
      'select exists(select 1 from %I.%I t '
      'where t.display_name = $1 and t.ns = $2 and t.uid <> $3)',
      TG_TABLE_SCHEMA, TG_TABLE_NAME
    ) into conflict_exists using dname, ns_val, NEW.uid;
  else
    execute format(
      'select exists(select 1 from %I.%I t '
      'where t.display_name = $1 and t.uid <> $2)',
      TG_TABLE_SCHEMA, TG_TABLE_NAME
    ) into conflict_exists using dname, NEW.uid;
  end if;

  if conflict_exists then
    raise exception 'display_name "%" is already used', dname
      using detail = 'SG0013',
            hint   = 'display_name must be unique within the namespace';
  end if;

  return NEW;
end;
$$ language plpgsql;

drop function if exists sgroups.check_rule_display_name_uniqueness_trg() cascade;
create or replace function sgroups.check_rule_display_name_uniqueness_trg()
returns trigger
as $$
declare
  conflict_exists boolean;
begin
  NEW.display_name := nullif(btrim(NEW.display_name::text), '')::sgroups.dname;
  if NEW.display_name is null then
    return NEW;
  end if;
  if NEW.ns is null then
    return NEW;
  end if;

  if TG_OP = 'UPDATE'
     and NEW.display_name is not distinct from OLD.display_name
     and NEW.ns           is not distinct from OLD.ns then
    return NEW;
  end if;

  select exists(
    select 1
      from sgroups.vu_rule_display_names v
     where v.display_name = NEW.display_name
       and v.ns           = NEW.ns
       and v.uid          <> NEW.uid
  ) into conflict_exists;

  if conflict_exists then
    raise exception 'display_name "%" is already used by another rule', NEW.display_name::text
      using detail = 'SG0013',
            hint   = 'display_name must be unique across all rules within the namespace';
  end if;

  return NEW;
end;
$$ language plpgsql;

do $$
declare
  t   text;
  trg text;
  tables_non_rule text[] := array[
    'sgroups.tbl_namespace',      'sgroups.tbl_ag',
    'sgroups.tbl_network',        'sgroups.tbl_host',
    'sgroups.tbl_host_binding',   'sgroups.tbl_network_binding',
    'sgroups.tbl_service',        'sgroups.tbl_service_binding'
  ];
  tables_rule text[] := array[
    'sgroups.tbl_ag2ag_rule',        'sgroups.tbl_ag2ag_icmp_rule',
    'sgroups.tbl_ag2icmp_rule',      'sgroups.tbl_ag2cidr_rule',
    'sgroups.tbl_ag2cidr_icmp_rule', 'sgroups.tbl_ag2fqdn_rule',
    'sgroups.tbl_svc2svc_rule',      'sgroups.tbl_svc2fqdn_rule',
    'sgroups.tbl_svc2cidr_rule',     'sgroups.tbl_svc2cidr_icmp_rule',
    'sgroups.tbl_ag2svc_rule',       'sgroups.tbl_svc2ag_rule',
    'sgroups.tbl_svc2ag_icmp_rule',  'sgroups.tbl_ag2svc_icmp_rule'
  ];
begin
  foreach t in array tables_non_rule loop
    trg := 'trg_' || replace(t, 'sgroups.tbl_', '') || '_display_name_uq';
    execute format('drop trigger if exists %I on %s', trg, t);
    execute format(
      'create trigger %I before insert or update on %s '
      'for each row execute function sgroups.check_display_name_uniqueness_trg()',
      trg, t
    );
  end loop;

  foreach t in array tables_rule loop
    trg := 'trg_' || replace(t, 'sgroups.tbl_', '') || '_display_name_uq';
    execute format('drop trigger if exists %I on %s', trg, t);
    execute format(
      'create trigger %I before insert or update on %s '
      'for each row execute function sgroups.check_rule_display_name_uniqueness_trg()',
      trg, t
    );
  end loop;
end$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

do $$
declare
  t   text;
  trg text;
  tables_all text[] := array[
    'sgroups.tbl_namespace',         'sgroups.tbl_ag',
    'sgroups.tbl_network',           'sgroups.tbl_host',
    'sgroups.tbl_host_binding',      'sgroups.tbl_network_binding',
    'sgroups.tbl_service',           'sgroups.tbl_service_binding',
    'sgroups.tbl_ag2ag_rule',        'sgroups.tbl_ag2ag_icmp_rule',
    'sgroups.tbl_ag2icmp_rule',      'sgroups.tbl_ag2cidr_rule',
    'sgroups.tbl_ag2cidr_icmp_rule', 'sgroups.tbl_ag2fqdn_rule',
    'sgroups.tbl_svc2svc_rule',      'sgroups.tbl_svc2fqdn_rule',
    'sgroups.tbl_svc2cidr_rule',     'sgroups.tbl_svc2cidr_icmp_rule',
    'sgroups.tbl_ag2svc_rule',       'sgroups.tbl_svc2ag_rule',
    'sgroups.tbl_svc2ag_icmp_rule',  'sgroups.tbl_ag2svc_icmp_rule'
  ];
begin
  foreach t in array tables_all loop
    trg := 'trg_' || replace(t, 'sgroups.tbl_', '') || '_display_name_uq';
    execute format('drop trigger if exists %I on %s', trg, t);
  end loop;
end$$;

drop function if exists sgroups.check_rule_display_name_uniqueness_trg() cascade;
drop function if exists sgroups.check_display_name_uniqueness_trg() cascade;
drop view if exists sgroups.vu_rule_display_names cascade;

alter table sgroups.tbl_namespace          drop constraint if exists ns_display_name_uq;
alter table sgroups.tbl_ag                 drop constraint if exists ag_display_name_uq;
alter table sgroups.tbl_network            drop constraint if exists network_display_name_uq;
alter table sgroups.tbl_host               drop constraint if exists host_display_name_uq;
alter table sgroups.tbl_host_binding       drop constraint if exists host_binding_display_name_uq;
alter table sgroups.tbl_network_binding    drop constraint if exists network_binding_display_name_uq;
alter table sgroups.tbl_service            drop constraint if exists service_display_name_uq;
alter table sgroups.tbl_service_binding    drop constraint if exists service_binding_display_name_uq;
alter table sgroups.tbl_ag2ag_rule         drop constraint if exists ag2ag_rule_display_name_uq;
alter table sgroups.tbl_ag2ag_icmp_rule    drop constraint if exists ag2ag_icmp_rule_display_name_uq;
alter table sgroups.tbl_ag2icmp_rule       drop constraint if exists ag2icmp_rule_display_name_uq;
alter table sgroups.tbl_ag2cidr_rule       drop constraint if exists ag2cidr_rule_display_name_uq;
alter table sgroups.tbl_ag2cidr_icmp_rule  drop constraint if exists ag2cidr_icmp_rule_display_name_uq;
alter table sgroups.tbl_ag2fqdn_rule       drop constraint if exists ag2fqdn_rule_display_name_uq;
alter table sgroups.tbl_svc2svc_rule       drop constraint if exists svc2svc_rule_display_name_uq;
alter table sgroups.tbl_svc2fqdn_rule      drop constraint if exists svc2fqdn_rule_display_name_uq;
alter table sgroups.tbl_svc2cidr_rule      drop constraint if exists svc2cidr_rule_display_name_uq;
alter table sgroups.tbl_svc2cidr_icmp_rule drop constraint if exists svc2cidr_icmp_rule_display_name_uq;
alter table sgroups.tbl_ag2svc_rule        drop constraint if exists ag2svc_rule_display_name_uq;
alter table sgroups.tbl_svc2ag_rule        drop constraint if exists svc2ag_rule_display_name_uq;
alter table sgroups.tbl_svc2ag_icmp_rule   drop constraint if exists svc2ag_icmp_rule_display_name_uq;
alter table sgroups.tbl_ag2svc_icmp_rule   drop constraint if exists ag2svc_icmp_rule_display_name_uq;

-- +goose StatementEnd
