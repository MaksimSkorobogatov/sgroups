-- +goose Up
-- +goose StatementBegin

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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd