BEGIN;
    INSERT INTO sgroups.tbl_namespace(name, uid, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'namespace-0',
            '368240f3-4866-4cbf-ad3f-6c55b48d4b92',
             '',
            '',
            '',
            'search => both',
            'search => both',
            '1'
        ),
        (
            'namespace-1',
            '82874cc8-0711-40fa-be42-dd480c4cb550',
            '',
            '',
            '',
            'search => labels',
            '',
            '1'
        ),
        (
            'namespace-2',
            '82040ed0-a028-4199-b266-3349096b5376',
            '',
            '',
            '',
            'labels => search',
            '',
            '1'
        ),
        (
            'namespace-3',
            'cd3e3c34-bf87-4787-87e8-7ba7182280c3',
            '',
            '',
            '',
            '',
            '',
            '1'
        ),
        (
            'namespace-4',
            'ccc9c066-2e2d-44aa-9c79-2b2ee0fabcc0',
            '',
            '',
            '',
            '',
            '',
            '1'
        ),
        (
            'namespace-5',
            '1e7c1f61-c021-4e53-8d53-79ee323264b9',
            '',
            '',
            '',
            '',
            '',
            '1'
        );

    INSERT INTO sgroups.tbl_ag(name, uid, ns, logs, trace, default_action, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'ag-0',
            'd459b92b-0881-4166-9035-6b994ebdf798',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            true,
            true,
            'DENY',
            'Address Group',
            'for search by name/ns',
            'address group for search',
            'labels => search',
            'search => labels',
            '1'
        ),
        (
            'ag-1',
            '37817691-8a8b-4344-8593-8a78ec3ce329',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            false,
            false,
            'ALLOW',
            'Address Group 1',
            'for search by name+ns',
            'address group for search',
            'labels => ns',
            'search => name',
            '1'
        ),
        (
            'ag-2',
            'ce6a67d4-c2fa-484f-ae74-85fcd9e65a21',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            false,
            true,
            'DENY',
            'Address Group 2',
            'for search by labels',
            'address group for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'ag-3',
            '9e9fb305-80bd-4748-ac42-fc208c398220',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            false,
            false,
            'ALLOW',
            'Address Group 3',
            'for edit',
            'address group for edit',
            'edit => ag',
            '',
            '1'
        ),
        (
            'ag-4',
            '596bbfe0-27ba-44f3-b460-99fab561d899',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            true,
            true,
            'ALLOW',
            'Address Group 4',
            'for delete ns+name',
            'address group for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'ag-5',
            '5cf8ec0d-5d62-472b-9287-6dbcedd87ead',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            true,
            true,
            'ALLOW',
            'Address Group 5',
            'for delete uid',
            'address group for delete',
            'delete => uid',
            '',
            '1'
        ),
        (
            'ag-6',
            'c8ad52a6-8440-4605-b90f-1441d11f3c79',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            true,
            true,
            'ALLOW',
            'Address Group 6',
            'for hb',
            'address group for host binding',
            'host => bind',
            '',
            '1'
        ),
        (
            'ag-7',
            'e225ec25-d4cf-42f8-8ea4-3adbcb02da49',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            true,
            true,
            'ALLOW',
            'Address Group 7',
            'for hb',
            'address group for host binding',
            'host => bind',
            '',
            '1'
        );

    INSERT INTO sgroups.tbl_host(name, uid, ns, ips, meta_info, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'host-0',
            'c38a4cb9-5689-41c2-9740-feeaa86a1b2e',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            '{192.168.1.1, 2001:db8::1}',
            '(host-0, null, null, null, null, null)',
            'host 0',
            'for search by name/ns',
            'host for search',
            'labels => search',
            'search => labels',
            '1'
        ),
        (
            'host-1',
            '013e25da-35bd-4e5f-b3de-5aa9c44564d0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '{127.0.0.1, ::1}',
            '(host-1, 1.0, null, null, null, null)',
            'host 1',
            'for search by name/ns',
            'host for search',
            'labels => ns',
            'search => name',
            '1'
        ),
        (
            'host-2',
            '150cb4f9-382b-44e1-b24e-28086f83bbc7',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '{5.5.5.5}',
            '(host-1, 1.0, linux, null, null, null)',
            'host 2',
            'for search by labels',
            'address group for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'host-3',
            'f315d9db-691e-4e9c-9bfc-546d59d35a68',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '{fe80::1}',
            '(host-1, 1.0, linux, ubuntu, null, null)',
            'host 3',
            'for search by labels',
            'address group for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'host-4',
            'f0ab4c7f-f2a2-4aed-854c-1e6d70b145e1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '{fe80::1}',
            '(host-1, 1.0, linux, ubuntu, family, null)',
            'host 4',
            'for delete ns+name',
            'address group for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'host-5',
            '62c5eb08-40d4-4f94-a766-70101ff7541f',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '{}',
            '(host-1, 1.0, linux, ubuntu, family, version)',
            'host 5',
            'for delete ns+name',
            'address group for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'host-6',
            '9fcce1ca-39ed-4ad0-8c1b-854497a1876f',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '{}',
            '(host-1, 1.0, linux, ubuntu, family, version)',
            'host 6',
            'for hb',
            'host for hb',
            'host => hb',
            '',
            '1'
        ),
        (
            'host-7',
            'd1d04af8-d1a2-44f9-b15d-41d5ffef20db',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '{}',
            '(host-1, 1.0, linux, ubuntu, family, version)',
            'host 7',
            'for hb',
            'host for hb',
            'host => hb',
            '',
            '1'
        );

    INSERT INTO sgroups.tbl_network(name, uid, ns, network, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'nw-0',
            'bd1fca5a-9c08-4ecd-9673-61ec60e9b88b',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            '1.1.1.1/32',
            'network 0',
            'for search by name/ns',
            'network for search',
            'labels => search',
            'search => labels',
            '1'
        ),
        (
            'nw-1',
            '987b4f77-0b5e-470e-85cf-037d79c37deb',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '::1/128',
            'network 1',
            'for search by name/ns',
            'network for search',
            'labels => ns',
            'search => name',
            '1'
        ),
        (
            'nw-2',
            'de8bad6a-6989-41fc-8af1-9ebc4c123f8c',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '10.0.0.0/8',
            'network 2',
            'for search by labels',
            'network for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'nw-3',
            'a55e2af7-1fae-4c2e-8429-79be384c7039',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '3.3.3.3/32',
            'network 3',
            'for search by labels',
            'network for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'nw-4',
            'cd94423b-fc15-4b5d-a029-05709f1af1fc',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '7.7.7.7/32',
            'network 4',
            'for delete ns+name',
            'network for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'nw-5',
            '45fceac3-1314-4ac8-b8c2-0fa6fd56cbb7',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '2.2.2.2/32',
            'network 5',
            'for delete ns+name',
            'network for delete',
            'network => name',
            '',
            '1'
        ),
        (
            'nw-6',
            '550e8400-e29b-41d4-a716-446655440000',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '2.2.2.2/32',
            'network 6',
            'for nb',
            'network for nb',
            'network => nb',
            '',
            '1'
        ),
        (
            'nw-7',
            '6ba7b810-9dad-11d1-80b4-00c04fd430c8',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '22.22.22.22/32',
            'network 7',
            'for nb',
            'network for nb',
            'network => nb',
            '',
            '1'
        );

    INSERT INTO sgroups.tbl_host_binding(name, uid, ns, ag, host, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'hb-0',
            '1085d231-9a0f-4697-b864-c0a522181911',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-0'),
            (SELECT id FROM sgroups.tbl_host WHERE name = 'host-0'),
            'host binding 0',
            'for search by name/ns',
            'host binding for search',
            'labels => search',
            'search => labels',
            '1'
        ),
        (
            'hb-1',
            'af7ee13a-ed99-4d35-bf68-ba76bb5a446c',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            (SELECT id FROM sgroups.tbl_host WHERE name = 'host-1'),
            'host binding 1',
            'for search by name/ns',
            'host binding for search',
            'labels => ns',
            'search => name',
            '1'
        ),
        (
            'hb-2',
            '5bd57191-447f-4c72-a295-4e8fe5800f20',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            (SELECT id FROM sgroups.tbl_host WHERE name = 'host-2'),
            'host binding 2',
            'for search by labels',
            'host binding for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'hb-3',
            'efb85252-c9b7-4dd5-b9be-6ce63276795c',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_host WHERE name = 'host-3'),
            'host binding 3',
            'for search by labels',
            'host binding for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'hb-4',
            'fa109b95-24ac-404a-9062-01ced3ab9541',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_host WHERE name = 'host-6'),
            'host binding 4',
            'for delete ns+name',
            'host binding for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'hb-5',
            '13155138-92ec-49e1-8a9e-2f1faac09dcd',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_host WHERE name = 'host-7'),
            'host binding 5',
            'for delete ns+name',
            'host binding for delete',
            'delete => name',
            '',
            '1'
        );

    INSERT INTO sgroups.tbl_network_binding(name, uid, ns, ag, network, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'nb-0',
            '90ef841a-f480-405d-abb0-983baf038801',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-0'),
            (SELECT id FROM sgroups.tbl_network WHERE name = 'nw-0'),
            'network binding 0',
            'for search by name/ns',
            'network binding for search',
            'labels => search',
            'search => labels',
            '1'
        ),
        (
            'nb-1',
            '9b187907-29ef-4de1-a798-d6537c8a95b1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            (SELECT id FROM sgroups.tbl_network WHERE name = 'nw-1'),
            'network binding 1',
            'for search by name/ns',
            'network binding for search',
            'labels => ns',
            'search => name',
            '1'
        ),
        (
            'nb-2',
            'fdcb56b7-94fe-412d-b7c7-e5ccf1f94085',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            (SELECT id FROM sgroups.tbl_network WHERE name = 'nw-2'),
            'network binding 2',
            'for search by labels',
            'network binding for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'nb-3',
            '507404f0-6ddc-4d15-b7bd-6c59bdff1d24',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_network WHERE name = 'nw-3'),
            'network binding 3',
            'for search by labels',
            'network binding for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'nb-4',
            '7019ef05-b363-42a5-a070-4d205a9d729e',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_network WHERE name = 'nw-6'),
            'network binding 4',
            'for delete ns+name',
            'network binding for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'nb-5',
            'c3aae183-6261-471b-a024-7c435319cfdb',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_network WHERE name = 'nw-7'),
            'network binding 5',
            'for delete ns+name',
            'network binding for delete',
            'delete => name',
            '',
            '1'
        );

INSERT INTO sgroups.tbl_service(
    name, uid, ns, transports,
    display_name, comment, description,
    labels, annotations, resource_version
)
VALUES
(
    'svc-0',
    'b1a3b2e4-6954-4688-af66-ab958988b416',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
    ARRAY[
        ROW(
            'udp',
            'IPv4',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[10000,10101),[10103,10104)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport,

        ROW(
            'tcp',
            'IPv4',
            ARRAY[
                ROW(
                    'some',
                    'some',
                    '{[10106,10110),[10115,10116)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 0',
    'for search by name/ns',
    'svc for search',
    'labels => search',
    'search => labels',
    '1'
),

(
    'svc-1',
    'c4451a8d-b0fa-4497-8efb-1d26f20fb2fa',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
    ARRAY[
        ROW(
            'icmp',
            'IPv4',
            ARRAY[
                ROW(
                    'combo',
                    'some',
                    NULL,
                    '{1, 2, 4, 18, 20}'
                )::sgroups.transport_entry
            ]
        )::sgroups.transport,

        ROW(
            'icmp',
            'IPv6',
            ARRAY[
                ROW(
                    'icmp types',
                    'some',
                    NULL,
                    '{255}'
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 1',
    'for search by name/ns',
    'svc for search',
    'labels => ns',
    'search => name',
    '1'
),

(
    'svc-2',
    '0f8009c1-16d9-400d-abad-6fe7c6d10cef',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
    ARRAY[
        ROW(
            'tcp',
            'IPv4',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[11000,11101),[11103,11104)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport,

        ROW(
            'icmp',
            'IPv4',
            ARRAY[
                ROW(
                    'types',
                    'some',
                    NULL,
                    '{0, 255}'
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 2',
    'for search by labels',
    'svc for search labels',
    'search => nameLabels',
    '',
    '1'
),

(
    'svc-3',
    '7fd507c2-f2f9-42cb-b7ab-e33b6a486faf',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
    ARRAY[
        ROW(
            'tcp',
            'IPv4',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[12000,12101),[12103,12104)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 3',
    'for search by labels',
    'svc for search labels',
    'search => nameLabels',
    '',
    '1'
),

(
    'svc-4',
    '3a151b99-7b90-4091-9b71-0441a0159377',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
    ARRAY[
        ROW(
            'udp',
            'IPv6',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[12000,12101),[12103,12104)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 4',
    'for delete ns+name',
    'svc for delete',
    'delete => name',
    '',
    '1'
),

(
    'svc-5',
    '3cf6cf63-27c0-48c9-928e-8bc820dafe71',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
    ARRAY[
        ROW(
            'tcp',
            'IPv4',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[12100,12101),[13103,13104)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 5',
    'for delete ns+name',
    'svc for delete',
    'svc => name',
    '',
    '1'
),

(
    'svc-6',
    '0ef21857-cd49-4c27-9398-7793f9d3a569',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
    ARRAY[
        ROW(
            'udp',
            'IPv4',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[14000,14101),[14103,14104)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 6',
    'for nb',
    'svc for nb',
    'svc => nb',
    '',
    '1'
),

(
    'svc-7',
    '85a610e8-3a7c-478f-9fe9-ce26b9e856b3',
    (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
    ARRAY[
        ROW(
            'tcp',
            'IPv4',
            ARRAY[
                ROW(
                    'ports',
                    'some',
                    '{[10106,10107)}',
                    NULL
                )::sgroups.transport_entry
            ]
        )::sgroups.transport
    ],
    'svc 7',
    'for nb',
    'svc for nb',
    'svc => nb',
    '',
    '1'
);

    INSERT INTO sgroups.tbl_service_binding(name, uid, ns, ag, service, display_name, comment, description, labels, annotations, resource_version)
    VALUES
        (
            'sb-0',
            'b724836e-0bbd-4cec-afa7-a8b8d45d4d6f',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-0'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-0'),
            'service binding 0',
            'for search by name/ns',
            'service binding for search',
            'labels => search',
            'search => labels',
            '1'
        ),
        (
            'sb-1',
            '7e13da25-2e49-4bae-bc5e-aa129870bc8a',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-1'),
            'service binding 1',
            'for search by name/ns',
            'service binding for search',
            'labels => ns',
            'search => name',
            '1'
        ),
        (
            'sb-2',
            '07f31279-b179-434c-ac45-65087d7dea91',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-2'),
            'service binding 2',
            'for search by labels',
            'network binding for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'sb-3',
            'a915bc00-2feb-42a5-94af-e209effbce44',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-3'),
            'service binding 3',
            'for search by labels',
            'service binding for search labels',
            'search => nameLabels',
            '',
            '1'
        ),
        (
            'sb-4',
            '5ce2863e-bde3-4c04-89d7-02a17144ee0b',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-6'),
            'service binding 4',
            'for delete ns+name',
            'service binding for delete',
            'delete => name',
            '',
            '1'
        ),
        (
            'sb-5',
            '44cee93f-db46-45df-b30e-3da74f4469d8',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-7'),
            'service binding 5',
            'for delete ns+name',
            'service binding for delete',
            'delete => name',
            '',
            '1'
        );

    INSERT INTO sgroups.tbl_rule_registry(name, ns, uid)
    VALUES
        (
            'rule-ag-ag-icmp-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            'dc7334ef-0c83-402e-b208-0f444e6fc674'
        ),
        (
            'rule-ag-ag-icmp-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '98af423d-a2fa-4da5-9b07-57d5c5379e16'
        ),
        (
            'rule-ag-ag-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'd9c9ef50-6a1e-4957-9484-770c4fc8c097'
        ),
        (
            'rule-ag-ag-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '483c720b-d926-45f4-80df-2d50310ace57'
        ),


        (
            'rule-ag-ag-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '2323fe5c-3148-47ed-95e6-d03ae2caec96'
        ),
        (
            'rule-ag-ag-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'ad45060e-e4fc-4ccb-8620-3ee4f44a3677'
        ),
        (
            'rule-ag-ag-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'e375da78-9d77-495e-91f1-12deaa0c7de3'
        ),
        (
            'rule-ag-ag-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '19208802-a15a-4234-9cc0-b5b0dc015e37'
        ),


        (
            'rule-ag-cidr-icmp-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '4d687030-6a77-47cf-8ac4-876b052eaa72'
        ),
        (
            'rule-ag-cidr-icmp-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '1d2ae37d-04d3-4f7b-988e-29f8eccb6944'
        ),
        (
            'rule-ag-cidr-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '733eea33-9191-4200-9dba-5dbb45b34d9e'
        ),
        (
            'rule-ag-cidr-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '567436a2-36a8-4ed4-b00c-eb062818c4ad'
        ),


        (
            'rule-ag-cidr-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '0e882827-7866-4cc6-b730-8a47140109ae'
        ),
        (
            'rule-ag-cidr-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '1fd5ac8b-b0d5-444d-878f-9dc4b22cb22d'
        ),
        (
            'rule-ag-cidr-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '4a070753-9de9-4516-8dce-fc32b54d7307'
        ),
        (
            'rule-ag-cidr-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '61c84179-5998-40da-98b4-5b3f3bdfa487'
        ),


        (
            'rule-ag-fqdn-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '94ee0a68-2f06-477c-9d58-197f84ba058f'
        ),
        (
            'rule-ag-fqdn-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '1befac7d-f817-47a7-a8de-809cbb078a54'
        ),
        (
            'rule-ag-fqdn-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'bad2d985-1f31-4880-86e0-e6a364b5c251'
        ),
        (
            'rule-ag-fqdn-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            'f381420a-7764-4f51-bfaa-a5e26155a001'
        ),


        (
            'rule-ag-icmp-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '78c75c27-a2f9-4150-ac9e-8cf42104f4de'
        ),
        (
            'rule-ag-icmp-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '9902905e-e2e4-4424-bd2b-e6fd303b3a0e'
        ),
        (
            'rule-ag-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '03966eca-c192-4b68-ba2c-d9801693020b'
        ),
        (
            'rule-ag-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '73268f3e-62eb-427d-8af7-7082eff8c774'
        ),


        (
            'rule-svc-cidr-icmp-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '0f0ba533-07a2-4eda-83a9-fd71e5f257cf'
        ),
        (
            'rule-svc-cidr-icmp-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '3798ac64-6e50-4e48-be78-24cb4cc8ce1c'
        ),
        (
            'rule-svc-cidr-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '64b6c20d-fc2c-4d01-acac-8a66ac5bcf36'
        ),
        (
            'rule-svc-cidr-icmp-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            'aa941eb3-9893-4290-98dc-ffd526b7cac6'
        ),


        (
            'rule-svc-cidr-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '12ce27dc-3789-4adb-8e76-d95e4145dfac'
        ),
        (
            'rule-svc-cidr-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            '90d8372b-e4d3-42e3-9502-9d35db9b6465'
        ),
        (
            'rule-svc-cidr-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'b267fcaf-b09a-4041-9437-c8d503258abc'
        ),
        (
            'rule-svc-cidr-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            '3af0aedf-2f4c-4122-baa5-4040f52470bf'
        ),


        (
            'rule-svc-fqdn-0',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'b9f791b0-7cd3-43b6-a17e-4b3ee7ff94ac'
        ),
        (
            'rule-svc-fqdn-1',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'dc95f2ab-0ac4-409c-91cc-40b742ead78c'
        ),
        (
            'rule-svc-fqdn-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            'c82c194f-a477-4af7-ac67-cc55f66a7c92'
        ),
        (
            'rule-svc-fqdn-2',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            'a25df54b-8c26-4bd9-a2c7-91115ee74095'
        );

    INSERT INTO sgroups.tbl_svc2fqdn_rule(name, uid, ns, svclocal, fqdn, display_name, comment, description, labels, annotations,  traffic, ip_v, action, proto, entries, resource_version)
    VALUES
        (
            'rule-svc-fqdn-0',
            'b9f791b0-7cd3-43b6-a17e-4b3ee7ff94ac',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-6'),
            'google.com',
            'rule svc-fqdn 0',
            'for search',
            'rule for search',
            'labels => search',
            'search => labels',
            'egress',
            'IPv4',
            'ALLOW',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[14000,14101),[14103,14104), [15000, 16475)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-svc-fqdn-1',
            'dc95f2ab-0ac4-409c-91cc-40b742ead78c',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-3'),
            'ya.ru',
            'rule svc-fqdn 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'egress',
            'IPv6',
            'ALLOW',
            'tcp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-svc-fqdn-2',
            'c82c194f-a477-4af7-ac67-cc55f66a7c92',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-1'),
            'twitter.com',
            'rule svc-fqdn 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-svc-fqdn-2',
            'a25df54b-8c26-4bd9-a2c7-91115ee74095',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-2'),
            'svc.cluster.local',
            'rule svc-fqdn 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_svc2cidr_rule(name, uid, ns, svclocal, cidr, display_name, comment, description, labels, annotations,  traffic, ip_v, action, proto, entries, resource_version)
    VALUES
        (
            'rule-svc-cidr-0',
            '12ce27dc-3789-4adb-8e76-d95e4145dfac',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-6'),
            '5.5.5.5/32',
            'rule svc-cidr 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[14000,14101),[14103,14104), [15000, 16475)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-svc-cidr-1',
            '90d8372b-e4d3-42e3-9502-9d35db9b6465',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-3'),
            '::/0',
            'rule svc-cidr 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'both',
            'IPv6',
            'ALLOW',
            'tcp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-svc-cidr-2',
            'b267fcaf-b09a-4041-9437-c8d503258abc',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-1'),
            '0.0.0.0/32',
            'rule svc-cidr 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'ingress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-svc-cidr-2',
            '3af0aedf-2f4c-4122-baa5-4040f52470bf',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-2'),
            '3.3.3.3/32',
            'rule svc-cidr 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_svc2cidr_icmp_rule(name, uid, ns, svclocal, cidr, display_name, comment, description, labels, annotations,  traffic, ip_v, action, entries, resource_version)
    VALUES
        (
            'rule-svc-cidr-icmp-0',
            '0f0ba533-07a2-4eda-83a9-fd71e5f257cf',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-6'),
            '::1/128',
            'rule svc-cidr-icmp 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-svc-cidr-icmp-1',
            '3798ac64-6e50-4e48-be78-24cb4cc8ce1c',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-1'),
            '10.0.0.0/8',
            'rule svc-cidr-icmp 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'egress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-svc-cidr-icmp-2',
            '64b6c20d-fc2c-4d01-acac-8a66ac5bcf36',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-1'),
            '2.2.2.2/32',
            'rule svc-cidr-icmp 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-svc-cidr-icmp-2',
            'aa941eb3-9893-4290-98dc-ffd526b7cac6',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_service WHERE name = 'svc-2'),
            '192.168.1.1/32',
            'rule svc-cidr-icmp 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_ag2icmp_rule(name, uid, ns, aglocal, display_name, comment, description, labels, annotations,  traffic, ip_v, action, entries, resource_version)
    VALUES
        (
            'rule-ag-icmp-0',
            '78c75c27-a2f9-4150-ac9e-8cf42104f4de',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            'rule ag-icmp 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-icmp-1',
            '9902905e-e2e4-4424-bd2b-e6fd303b3a0e',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            'rule ag-icmp 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'egress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-icmp-2',
            '03966eca-c192-4b68-ba2c-d9801693020b',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            'rule ag-icmp 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-icmp-2',
            '73268f3e-62eb-427d-8af7-7082eff8c774',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            'rule ag-icmp 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_ag2fqdn_rule(name, uid, ns, aglocal, fqdn, display_name, comment, description, labels, annotations,  traffic, ip_v, action, proto, entries, resource_version)
    VALUES
        (
            'rule-ag-fqdn-0',
            '94ee0a68-2f06-477c-9d58-197f84ba058f',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-6'),
            'google.com',
            'rule ag-fqdn 0',
            'for search',
            'rule for search',
            'labels => search',
            'search => labels',
            'egress',
            'IPv4',
            'ALLOW',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[14000,14101),[14103,14104), [15000, 16475)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-fqdn-1',
            '1befac7d-f817-47a7-a8de-809cbb078a54',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            'ya.ru',
            'rule ag-fqdn 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'egress',
            'IPv6',
            'ALLOW',
            'tcp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-fqdn-2',
            'bad2d985-1f31-4880-86e0-e6a364b5c251',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            'twitter.com',
            'rule ag-fqdn 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-fqdn-2',
            'f381420a-7764-4f51-bfaa-a5e26155a001',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            'svc.cluster.local',
            'rule ag-fqdn 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_ag2cidr_rule(name, uid, ns, aglocal, cidr, display_name, comment, description, labels, annotations,  traffic, ip_v, action, proto, entries, resource_version)
    VALUES
        (
            'rule-ag-cidr-0',
            '0e882827-7866-4cc6-b730-8a47140109ae',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-6'),
            '5.5.5.5/32',
            'rule ag-cidr 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[14000,14101),[14103,14104), [15000, 16475)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-cidr-1',
            '1fd5ac8b-b0d5-444d-878f-9dc4b22cb22d',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            '::/0',
            'rule ag-cidr 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'both',
            'IPv6',
            'ALLOW',
            'tcp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-cidr-2',
            '4a070753-9de9-4516-8dce-fc32b54d7307',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            '0.0.0.0/32',
            'rule ag-cidr 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'ingress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-cidr-2',
            '61c84179-5998-40da-98b4-5b3f3bdfa487',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            '3.3.3.3/32',
            'rule ag-cidr 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_ag2cidr_icmp_rule(name, uid, ns, aglocal, cidr, display_name, comment, description, labels, annotations,  traffic, ip_v, action, entries, resource_version)
    VALUES
        (
            'rule-ag-cidr-icmp-0',
            '4d687030-6a77-47cf-8ac4-876b052eaa72',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-6'),
            '::1/128',
            'rule ag-cidr-icmp 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-cidr-icmp-1',
            '1d2ae37d-04d3-4f7b-988e-29f8eccb6944',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            '10.0.0.0/8',
            'rule ag-cidr-icmp 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'egress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-cidr-icmp-2',
            '733eea33-9191-4200-9dba-5dbb45b34d9e',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            '2.2.2.2/32',
            'rule ag-cidr-icmp 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-cidr-icmp-2',
            '567436a2-36a8-4ed4-b00c-eb062818c4ad',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            '192.168.1.1/32',
            'rule ag-cidr-icmp 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_ag2ag_icmp_rule(name, uid, ns, aglocal, agremote, display_name, comment, description, labels, annotations,  traffic, ip_v, action, entries, resource_version)
    VALUES
        (
            'rule-ag-ag-icmp-0',
            'dc7334ef-0c83-402e-b208-0f444e6fc674',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-0'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-0'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-0'),
            'rule ag-ag-icmp 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-ag-icmp-1',
            '98af423d-a2fa-4da5-9b07-57d5c5379e16',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            'rule ag-ag-icmp 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'egress',
            'IPv4',
            'ALLOW',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,255}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-ag-icmp-2',
            'd9c9ef50-6a1e-4957-9484-770c4fc8c097',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            'rule ag-ag-icmp 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        ),
        (
            'rule-ag-ag-icmp-2',
            '483c720b-d926-45f4-80df-2d50310ace57',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            'rule ag-ag-icmp 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
             ARRAY[
                ROW(
                    'some description',
                    'some comment',
                    '{0,1,2}'::sgroups.icmp_types
                )::sgroups.icmp_entries
            ],
            '1'
        );

    INSERT INTO sgroups.tbl_ag2ag_rule(name, uid, ns, aglocal, agremote, display_name, comment, description, labels, annotations,  traffic, ip_v, action, proto, entries, resource_version)
    VALUES
        (
            'rule-ag-ag-0',
            '2323fe5c-3148-47ed-95e6-d03ae2caec96',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-6'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-7'),
            'rule ag-ag 0',
            'for search',
            'service binding for search',
            'labels => search',
            'search => labels',
            'ingress',
            'IPv4',
            'ALLOW',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[14000,14101),[14103,14104), [15000, 16475)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-ag-1',
            'ad45060e-e4fc-4ccb-8620-3ee4f44a3677',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            'rule ag-ag 1',
            'for edit',
            'rule for edit',
            'for => edit',
            'edit => rule',
            'both',
            'IPv6',
            'ALLOW',
            'tcp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-ag-2',
            'e375da78-9d77-495e-91f1-12deaa0c7de3',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-1'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            'rule ag-ag 2',
            'for delete',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'ingress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        ),
        (
            'rule-ag-ag-2',
            '19208802-a15a-4234-9cc0-b5b0dc015e37',
            (SELECT id FROM sgroups.tbl_namespace WHERE name = 'namespace-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-2'),
            (SELECT id FROM sgroups.tbl_ag WHERE name = 'ag-3'),
            'rule ag-ag 2',
            'for delete by uid',
            'rule for delete',
            'for => delete',
            'delete => rule',
            'egress',
            'IPv6',
            'DENY',
            'udp',
             ARRAY[
                ROW(
                    'ports 1',
                    'some comment',
                    '{[13000,13101),[13103,13104), [14000, 15475)}'::sgroups.port_ranges
                )::sgroups.port_entries,
                 ROW(
                    'ports 2',
                    'some comment',
                    '{[7000,7101),[8103,9104)}'::sgroups.port_ranges
                )::sgroups.port_entries
            ],
            '1'
        );
COMMIT;