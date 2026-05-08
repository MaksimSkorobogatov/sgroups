-- +goose Up
-- +goose StatementBegin

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

-- +goose StatementEnd
