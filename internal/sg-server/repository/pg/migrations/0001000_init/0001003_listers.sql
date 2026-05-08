-- +goose Up
-- +goose StatementBegin

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


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd