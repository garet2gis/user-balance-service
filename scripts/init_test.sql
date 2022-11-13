CREATE TABLE balance
(
    user_id UUID PRIMARY KEY,
    balance decimal(18, 2) NOT NULL CHECK ( balance >= 0 ) DEFAULT 0
);

CREATE TABLE service
(
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE reservation
(
    reservation_id UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id        UUID           NOT NULL,
    order_id       UUID           NOT NULL,
    service_id     UUID           NOT NULL,
    cost           decimal(18, 2) NOT NULL CHECK ( cost > 0 ),
    created_at     TIMESTAMP      NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    comment        TEXT           NOT NULL DEFAULT '',

    CONSTRAINT uq_reservation UNIQUE (user_id, order_id, service_id, cost),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES balance (user_id),

    CONSTRAINT fk_service
        FOREIGN KEY (service_id)
            REFERENCES service (service_id)
);

CREATE TYPE reservation_status AS ENUM ('confirm', 'cancel');
CREATE TABLE history_reservation
(
    commit_reservation_id UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    user_id               UUID               NOT NULL,
    order_id              UUID               NOT NULL,
    service_id            UUID               NOT NULL,
    cost                  decimal(18, 2)
        CHECK (cost <> 0 AND (cost < 0 OR status = 'cancel') AND (cost > 0 OR status = 'confirm'))
                                             NOT NULL,
    comment               TEXT               NOT NULL DEFAULT '',
    status                reservation_status NOT NULL,
    created_at            TIMESTAMP          NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);


CREATE TABLE history_deposit
(
    history_deposit_id UUID PRIMARY KEY                                                  DEFAULT gen_random_uuid(),
    user_id            UUID                               NOT NULL,
    from_user_id       UUID CHECK (from_user_id <> user_id)                              DEFAULT NULL,
    to_user_id         UUID CHECK (to_user_id <> user_id AND to_user_id <> from_user_id) DEFAULT NULL,
    amount             decimal(18, 2) CHECK (amount <> 0) NOT NULL,
    comment            TEXT                               NOT NULL                       DEFAULT '',
    created_at         TIMESTAMP                          NOT NULL                       DEFAULT (now() AT TIME ZONE 'utc'),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES balance (user_id),
    CONSTRAINT fk_from_user
        FOREIGN KEY (from_user_id)
            REFERENCES balance (user_id)
);


CREATE VIEW balance_history AS
SELECT reservation.user_id,
       CAST(NULL AS UUID)     as from_user_id,
       CAST(NULL AS UUID)     as to_user_id,
       reservation.order_id,
       service.name           as service_name,
       reservation.created_at as create_date,
       reservation.cost       as amount,
       reservation.comment,
       'reserve'              as transaction_type
FROM reservation
         JOIN service USING (service_id)

UNION

SELECT history_reservation.user_id,
       CAST(NULL AS UUID)                      as from_user_id,
       CAST(NULL AS UUID)                      as to_user_id,
       history_reservation.order_id,
       service.name                            as service_name,
       history_reservation.created_at          as create_date,
       history_reservation.cost                as amount,
       history_reservation.comment,
       history_reservation.status::varchar(32) as transaction_type
FROM history_reservation
         JOIN service USING (service_id)

UNION

SELECT history_deposit.user_id,
       history_deposit.from_user_id,
       history_deposit.to_user_id,
       CAST(NULL AS UUID)         as order_id,
       ''                         as service_name,
       history_deposit.created_at as create_date,
       history_deposit.amount,
       history_deposit.comment,
       'balance_change'           as transaction_type
FROM history_deposit;

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