-- +goose Up
-- +goose StatementBegin

-- list_ag with fast-path for simple selectors. vu_ag.refs is the heaviest
-- (host + network + service + 5 rule-ref helpers per row), so skipping its
-- materialization for non-matching rows gives the biggest win.
drop function if exists sgroups.list_ag(sgroups.res_selector[]) cascade;
create or replace function sgroups.list_ag(
  selectors sgroups.res_selector[] default null
)
returns setof sgroups.vu_ag
as $$
declare
  has_complex   boolean;
  has_match_all boolean;
begin
  if selectors is null or cardinality(selectors) = 0 then
    return query select * from sgroups.vu_ag;
    return;
  end if;

  select
    bool_or(
      (s.label_selector is not null
        and cardinality(akeys(s.label_selector)) > 0)
      or (
        (s.field_selector).refs is not null
        and cardinality((s.field_selector).refs) > 0
      )
    ),
    bool_or(
      nullif(btrim((s.field_selector).name), '') is null
      and nullif(btrim((s.field_selector).namespace), '') is null
      and (s.label_selector is null
           or cardinality(akeys(s.label_selector)) = 0)
      and (
        (s.field_selector).refs is null
        or cardinality((s.field_selector).refs) = 0
      )
    )
  into has_complex, has_match_all
  from unnest(selectors) s;

  if has_match_all then
    return query select * from sgroups.vu_ag;
    return;
  end if;

  if has_complex then
    return query
    select v.*
    from sgroups.vu_ag v
    where sgroups.match_res_selectors(
      sgroups.mk_res_selector(v.name, v.namespace, v.refs, v.labels),
      selectors
    );
    return;
  end if;

  return query
  select v.*
  from sgroups.vu_ag v
  where exists (
    select 1
    from unnest(selectors) s
    where (
        nullif(btrim((s.field_selector).name), '') is null
        or nullif(btrim((s.field_selector).name), '') = v.name
      )
      and (
        nullif(btrim((s.field_selector).namespace), '') is null
        or nullif(btrim((s.field_selector).namespace), '') = v.namespace
      )
  );
end;
$$ language plpgsql stable;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
