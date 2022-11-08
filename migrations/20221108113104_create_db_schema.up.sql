CREATE TABLE balance
(
    user_id UUID PRIMARY KEY,
    balance decimal(18, 2) NOT NULL CHECK ( balance > 0 )
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

    CONSTRAINT uq_reservation UNIQUE (user_id, order_id, service_id, cost),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES balance (user_id),

    CONSTRAINT fk_service
        FOREIGN KEY (service_id)
            REFERENCES service (service_id)
);

CREATE TYPE reservation_status AS ENUM ('confirm', 'cancel');
CREATE TABLE commit_reservation
(
    commit_reservation_id UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    user_id               UUID               NOT NULL,
    order_id              UUID               NOT NULL,
    service_id            UUID               NOT NULL,
    cost                  decimal(18, 2)     NOT NULL CHECK ( cost > 0 ),
    status                reservation_status NOT NULL,
    created_at            TIMESTAMP          NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);


CREATE TABLE replenishment
(
    replenishment_id UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    user_id    UUID           NOT NULL,
    amount     decimal(18, 2) NOT NULL CHECK ( amount > 0 ),
    created_at TIMESTAMP      NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),

    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES balance (user_id)
);


CREATE VIEW balance_history AS
SELECT reservation.user_id,
       reservation.order_id,
       reservation.service_id,
       reservation.created_at as create_date,
       reservation.cost       as amount,
       'reserve'              as transaction_type
FROM reservation

UNION

SELECT commit_reservation.user_id,
       commit_reservation.order_id,
       commit_reservation.service_id,
       commit_reservation.created_at          as create_date,
       commit_reservation.cost                as amount,
       commit_reservation.status::varchar(32) as transaction_type
FROM commit_reservation

UNION

SELECT replenishment.user_id,
       NULL                     as order_id,
       NULL                     as service_id,
       replenishment.created_at as create_date,
       replenishment.amount,
       'replenish'              as transaction_type
FROM replenishment;