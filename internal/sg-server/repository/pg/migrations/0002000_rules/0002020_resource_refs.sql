-- +goose Up
-- +goose StatementBegin

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
