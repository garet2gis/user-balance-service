INSERT INTO balance (user_id, balance)
VALUES ('7a13445c-d6df-4111-abc0-abb12f610069', 500.34),
       ('5bb0d661-cf53-419a-93ec-a9d8938b2d54', 200),
       ('7a270640-c4eb-414d-86e1-b8ff0c6faf6e', 300),
       ('028abfaa-f28c-43f2-8fec-97010fda1080', 100),
       ('6dddad21-fb69-4532-bd7d-da861548159f', 100),
       ('21a09284-440e-45fb-ab63-183fce3bd43c', 100),
       ('9125aaa5-2167-484c-ab38-2efc597c3405', 100),
       ('631d7ad9-413b-4253-9bd7-ecabd21555fb', 100),
       ('f3289ed7-b57a-4b27-ab3f-f52ba690f861', 100),
       ('983e8792-6736-41bd-9f1a-7c67f8501645', 100),
       ('17dc1e89-37c0-4e77-9770-b83e82aa923d', 100),
       ('bf13b3f8-503d-4e41-8f71-a541a20583e6', 100),
       ('1e472747-8ccf-4fef-9d65-2fdc71a72568', 100),
       ('1994126f-9fa6-4334-a9f2-df47a83679ad', 100),
       ('099e51b7-0ee9-406c-87d9-8bdec0a9b527', 100),
       ('12bbc290-ff27-4da6-9e15-165cba326dc8', 100),
       ('619b2cad-d087-4cf6-8b3c-460f468f46be', 100),
       ('232daa0f-aa23-4cff-8fbf-380730e7f286', 100),
       ('b55e4e01-5152-4cb0-95f2-ee27d5d2e9cd', 100),
       ('a1a2c822-9a7c-4488-8c0d-bc858088460d', 100),
       ('b0ca3505-4d5a-4120-85bf-8610b58c8678', 100),
       ('bb79ef66-fead-4d6d-a32e-1e00c765353a', 100),
       ('abc0cb36-5e18-4138-ad38-d7b2c1f998d1', 100),
       ('ecc9412a-f79e-4e97-ae3d-7180d3750b2f', 100);


INSERT INTO service (service_id, name)
VALUES ('34e16535-480c-43f8-95a9-b7a503499afd', 'Услуга 1');

INSERT INTO reservation (user_id, order_id, service_id, cost)
VALUES ('7a13445c-d6df-4111-abc0-abb12f610069', '34e16535-480c-43f8-95a9-b7a503499afb',
        '34e16535-480c-43f8-95a9-b7a503499afd', 50),
       ('7a13445c-d6df-4111-abc0-abb12f610069', 'b55e4e01-5152-4cb0-95f2-ee27d5d2e9cd',
        '34e16535-480c-43f8-95a9-b7a503499afd', 100);

INSERT INTO history_reservation (user_id, order_id, service_id, cost, status)
VALUES ('7a13445c-d6df-4111-abc0-abb12f610069', '34e16535-480c-43f8-95a9-b7a503499afb',
        '34e16535-480c-43f8-95a9-b7a503499afd', 100, 'cancel'),
       ('7a13445c-d6df-4111-abc0-abb12f610069', '34e16535-480c-43f8-95a9-b7a503499afb',
        '34e16535-480c-43f8-95a9-b7a503499afd', -70, 'confirm');



-- INSERT INTO replenishment (user_id, amount)
-- VALUES ('7a13445c-d6df-4111-abc0-abb12f610069', 80),
--        ('7a13445c-d6df-4111-abc0-abb12f610069', 80),
--        ('1e472747-8ccf-4fef-9d65-2fdc71a72568', 80);

SELECT *
FROM balance_history;

SELECT service.name, SUM(-history_reservation.cost) as "sum"
FROM history_reservation
         JOIN service USING (service_id)
WHERE history_reservation.status = 'confirm'
  AND EXTRACT(YEAR FROM history_reservation.created_at) = 2022
  AND EXTRACT(MONTH FROM history_reservation.created_at) = 11
GROUP BY service.name;


SELECT balance_history.order_id,
       balance_history.service_name,
       balance_history.from_user_id,
       balance_history.create_date,
       balance_history.amount,
       balance_history.transaction_type,
       balance_history.comment
FROM balance_history
WHERE balance_history.user_id = '';


DELETE
FROM reservation
WHERE user_id = ''
  AND order_id = ''
  AND service_id = ''
  AND cost = 0