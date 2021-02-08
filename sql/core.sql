-- Init
CREATE DATABASE wallmask;
CREATE TABLE proxies
(
    id         SMALLSERIAL PRIMARY KEY,
    protocol   TEXT NOT NULL,
    ipv4       TEXT NOT NULL,
    port       INT  NOT NULL,
    lastTested TIMESTAMP,
    working    BOOL NOT NULL
);
SELECT *
FROM proxies;
SELECT COUNT(*)
FROM proxies;
DROP TABLE proxies;

CREATE USER wallmaskcli WITH PASSWORD 'a1gmgs8XeO1KW1kG6xK6LS2DKwqAVqEguWy7jlmhFnUKBVeehakVxLf25h';


-- Specifics

SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public';

SELECT COUNT(*)
FROM proxies;

SELECT lastTested, protocol || '://' || ipv4 || ':' || CAST(port AS TEXT) addr
FROM proxies
WHERE working
  AND protocol != ''
  AND ipv4 != ''
ORDER BY lastTested DESC;

SELECT id
FROM proxies
WHERE ipv4 = ''
  AND port = 0;

INSERT INTO proxies (ipv4, port, lastTested, working)
VALUES ('$1', 0, 1609018536968587000, true);


-- TEST
CREATE TABLE test
(
    id     SMALLSERIAL PRIMARY KEY,
    n BIGINT NOT NULL
);
DROP TABLE test;
INSERT INTO test (n)
VALUES (123);
SELECT *
FROM test;
SELECT COUNT(*)
FROM test;