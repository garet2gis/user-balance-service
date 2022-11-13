INSERT INTO balance (user_id, balance)
VALUES ('7a13445c-d6df-4111-abc0-abb12f610069', 500.34),
       ('7a13445c-d6df-4111-abc0-abb12f610068', 121),
       ('7a13445c-d6df-4111-abc0-abb12f610062', 32.32),
       ('7a13445c-d6df-4111-abc0-abb12f610063', 0),
       ('7a13445c-d6df-4111-abc0-abb12f610064', 0),
       ('7a13445c-d6df-4111-abc0-abb12f610065', 0);

INSERT INTO service (service_id, name)
VALUES ('34e16535-480c-43f8-95a9-b7a503499af0', 'Курьерская доставка'),
       ('34e16535-480c-43f8-95a9-b7a503499af1', 'Бронирование'),
       ('34e16535-480c-43f8-95a9-b7a503499af2', 'Дополнительная гарантия для товара');


INSERT INTO reservation (user_id, order_id, service_id, cost, comment)
VALUES ('7a13445c-d6df-4111-abc0-abb12f610068',
        '983e8792-6736-41bd-9f1a-7c67f8501645',
        '34e16535-480c-43f8-95a9-b7a503499af2',
        50,
        'reserve 50');

INSERT INTO history_reservation (user_id, order_id, service_id, cost, status)
VALUES ('7a13445c-d6df-4111-abc0-abb12f610065',
        '983e8792-6736-41bd-9f1a-7c67f8501645',
        '34e16535-480c-43f8-95a9-b7a503499af2',
        -50.34,
        'confirm'),
       ('7a13445c-d6df-4111-abc0-abb12f610065',
        '983e8792-6736-41bd-9f1a-7c67f8501645',
        '34e16535-480c-43f8-95a9-b7a503499af2',
        -20.40,
        'confirm'),
       ('7a13445c-d6df-4111-abc0-abb12f610065',
        '983e8792-6736-41bd-9f1a-7c67f8501645',
        '34e16535-480c-43f8-95a9-b7a503499af0',
        -120.78,
        'confirm'),
       ('7a13445c-d6df-4111-abc0-abb12f610065',
        '983e8792-6736-41bd-9f1a-7c67f8501645',
        '34e16535-480c-43f8-95a9-b7a503499af1',
        -57,
        'confirm'),
       ('7a13445c-d6df-4111-abc0-abb12f610065',
        '983e8792-6736-41bd-9f1a-7c67f8501645',
        '34e16535-480c-43f8-95a9-b7a503499af1',
        7,
        'cancel')
;