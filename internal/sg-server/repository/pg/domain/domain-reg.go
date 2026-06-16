package domain

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

// RegisterSGroupsTypesOntoPGX registers sgroups domain types onto PGX lib
func RegisterSGroupsTypesOntoPGX(ctx context.Context, c *pgx.Conn) (err error) {
	type reg struct {
		typeName string
		run      func(string) error
	}

	regs := [...]reg{
		{
			typeName: "sgroups.rname,sgroups.dname,sgroups.sync_op,sgroups.resource_op",
			run: func(tn string) (err error) {
				for _, t := range strings.Split(tn, ",") {
					if err = regPgType(ctx, c, t); err != nil {
						err = errors.WithMessagef(err, "register PG type for '%s' typename", t)
						break
					}
				}
				return err
			},
		},
		{
			typeName: "citext",
			run: func(tn string) error {
				return regPgCitextType(ctx, c)
			},
		},
		{
			typeName: "hstore",
			run: func(tn string) error {
				return regPgHstoreType(ctx, c, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[map[string]string](c, tElem.Name)
					regDefPgType[*map[string]string](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.fqdn",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[FQDN](c, tElem.Name)
					regDefPgType[*FQDN](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.port_ranges",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[PortMultirange](c, tElem.Name)
					regDefPgType[*PortMultirange](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.policy_action",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[PolicyAction](c, tElem.Name)
					regDefPgType[*PolicyAction](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.traffic",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[Traffic](c, tElem.Name)
					regDefPgType[*Traffic](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.proto",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[Proto](c, tElem.Name)
					regDefPgType[*Proto](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.ip_family",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[IpFamily](c, tElem.Name)
					regDefPgType[*IpFamily](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.icmp_types",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[IcmpTypes](c, tElem.Name)
					regDefPgType[*IcmpTypes](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.resource_type",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[ResourceType](c, tElem.Name)
					regDefPgType[*ResourceType](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.host_info",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[HostInfo](c, tElem.Name)
					regDefPgType[*HostInfo](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.named_port",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[NamedPort](c, tElem.Name)
					regDefPgType[*NamedPort](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.host_endpoints",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[HostEndpoints](c, tElem.Name)
					regDefPgType[*HostEndpoints](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.resource_id",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[ResourceIdentifier](c, tElem.Name)
					regDefPgType[*ResourceIdentifier](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.resource_ref",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[ResourceRef](c, tElem.Name)
					regDefPgType[*ResourceRef](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.field_selector",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[FieldSelector](c, tElem.Name)
					regDefPgType[*FieldSelector](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.res_selector",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[ResSelector](c, tElem.Name)
					regDefPgType[*ResSelector](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.endpoint",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[Endpoint](c, tElem.Name)
					regDefPgType[*Endpoint](c, tElem.Name)
					return nil
				})
			},
		},
		{
			typeName: "sgroups.port_entries",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[PortEntries](c, tElem.Name)
					regDefPgType[*PortEntries](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.icmp_entries",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[IcmpEntries](c, tElem.Name)
					regDefPgType[*IcmpEntries](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.transport_entry",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[TransportEntry](c, tElem.Name)
					regDefPgType[*TransportEntry](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
		{
			typeName: "sgroups.transport",
			run: func(tn string) error {
				return regPgType(ctx, c, tn, func(c *pgx.Conn, tElem *pgtype.Type) error {
					regDefPgType[Transport](c, tElem.Name)
					regDefPgType[*Transport](c, tElem.Name)
					return regPgArrayType(ctx, c, tElem)
				})
			},
		},
	}

	for _, r := range regs {
		if err = r.run(r.typeName); err != nil {
			err = errors.WithMessage(err, r.typeName)
			break
		}
	}

	return errors.WithMessage(err, "register 'sgroups' types onto PGX")
}

func regPgType(ctx context.Context, c *pgx.Conn, pgTypename string, onOk ...func(*pgx.Conn, *pgtype.Type) error) error {
	pgType, err := c.LoadType(ctx, pgTypename)
	if err != nil {
		return errors.WithMessagef(err, "load PG type for '%s' typename", pgTypename)
	}
	tm := c.TypeMap()
	tm.RegisterType(pgType)
	for _, f := range onOk {
		if err = f(c, pgType); err != nil {
			break
		}
	}
	return err
}

func regPgArrayType(ctx context.Context, c *pgx.Conn, pgElemType *pgtype.Type, onOk ...func(*pgx.Conn, *pgtype.Type) error) error { //nolint:unparam
	arrayPgType := &pgtype.Type{
		Codec: &pgtype.ArrayCodec{ElementType: pgElemType},
	}
	err := c.QueryRow(ctx, "select oid, typname from pg_type where typelem=$1", pgElemType.OID).
		Scan(
			&arrayPgType.OID,
			&arrayPgType.Name,
		)
	if err != nil {
		return errors.WithMessagef(err, "find 'OID' and 'Typename' for PG array-of['%s']", pgElemType.Name)
	}
	tm := c.TypeMap()
	tm.RegisterType(arrayPgType)
	for _, f := range onOk {
		if err = f(c, arrayPgType); err != nil {
			break
		}
	}
	return err
}

func regDefPgType[t any](c *pgx.Conn, pgTypename string) {
	tm := c.TypeMap()
	var x t
	tm.RegisterDefaultPgType(x, pgTypename)
}

func regPgHstoreType(ctx context.Context, c *pgx.Conn, onOk ...func(*pgx.Conn, *pgtype.Type) error) error {
	var hstoreOID uint32
	if err := c.QueryRow(ctx, "select $1::text::regtype::oid", "public.hstore").Scan(&hstoreOID); err != nil {
		return errors.WithMessage(err, "lookup OID for 'hstore' via regtype")
	}

	hstoreType := &pgtype.Type{
		Name:  "hstore",
		OID:   hstoreOID,
		Codec: pgtype.HstoreCodec{},
	}
	tm := c.TypeMap()
	tm.RegisterType(hstoreType)

	var hstoreArrayOID uint32
	if err := c.QueryRow(ctx, "select $1::text::regtype::oid", "public.hstore[]").Scan(&hstoreArrayOID); err == nil && hstoreArrayOID != 0 {
		tm.RegisterType(&pgtype.Type{
			Name:  "hstore[]",
			OID:   hstoreArrayOID,
			Codec: &pgtype.ArrayCodec{ElementType: hstoreType},
		})
	}

	for _, f := range onOk {
		if err := f(c, hstoreType); err != nil {
			return err
		}
	}
	return nil
}

func regPgCitextType(ctx context.Context, c *pgx.Conn, onOk ...func(*pgx.Conn, *pgtype.Type) error) error {
	var citextOID uint32
	if err := c.QueryRow(ctx, "select $1::text::regtype::oid", "citext").Scan(&citextOID); err != nil {
		return errors.WithMessage(err, "lookup OID for 'citext' via regtype (ensure extension citext is installed)")
	}

	citextType := &pgtype.Type{
		Name:  "citext",
		OID:   citextOID,
		Codec: pgtype.TextCodec{},
	}
	tm := c.TypeMap()
	tm.RegisterType(citextType)

	var citextArrayOID uint32
	if err := c.QueryRow(ctx, "select $1::text::regtype::oid", "citext[]").Scan(&citextArrayOID); err == nil && citextArrayOID != 0 {
		tm.RegisterType(&pgtype.Type{
			Name:  "citext[]",
			OID:   citextArrayOID,
			Codec: &pgtype.ArrayCodec{ElementType: citextType},
		})
	}

	for _, f := range onOk {
		if err := f(c, citextType); err != nil {
			return err
		}
	}
	return nil
}
