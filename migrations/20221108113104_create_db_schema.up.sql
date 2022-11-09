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

CREATE TYPE reservation_status AS ENUM ('confirm', 'cancel', 'was_reserve');
CREATE TABLE history_reservation
(
    commit_reservation_id UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    user_id               UUID               NOT NULL,
    order_id              UUID               NOT NULL,
    service_id            UUID               NOT NULL,
    cost                  decimal(18, 2)     NOT NULL,
    comment               TEXT               NOT NULL DEFAULT '',
    status                reservation_status NOT NULL,
    created_at            TIMESTAMP          NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);


CREATE TABLE replenishment
(
    replenishment_id UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id          UUID           NOT NULL,
    amount           decimal(18, 2) NOT NULL CHECK ( amount > 0 ),
    comment          TEXT           NOT NULL DEFAULT '',
    created_at       TIMESTAMP      NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES balance (user_id)
);


CREATE VIEW balance_history AS
SELECT reservation.user_id,
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
       history_reservation.order_id,
       service.name                           as service_name,
       history_reservation.created_at          as create_date,
       history_reservation.cost                as amount,
       history_reservation.comment,
       history_reservation.status::varchar(32) as transaction_type
FROM history_reservation
         JOIN service USING (service_id)

UNION

SELECT replenishment.user_id,
       NULL                     as order_id,
       ''                       as service_name,
       replenishment.created_at as create_date,
       replenishment.amount,
       replenishment.comment,
       'replenish'              as transaction_type
FROM replenishment;