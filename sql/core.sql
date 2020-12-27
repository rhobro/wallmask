-- Init
CREATE DATABASE wallmask;
CREATE TABLE proxies
(
    id         SMALLSERIAL PRIMARY KEY,
    ipv4       TEXT NOT NULL,
    port       INT,
    lastTested TIMESTAMP,
    working    BOOL NOT NULL
);
DROP TABLE proxies;
SELECT *
FROM proxies;
SELECT COUNT(*)
FROM proxies;

CREATE USER wallmaskcli WITH PASSWORD 'a1gmgs8XeO1KW1kG6xK6LS2DKwqAVqEguWy7jlmhFnUKBVeehakVxLf25h';


-- Specifics
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public';

SELECT *
FROM proxies
WHERE working
ORDER BY lastTested DESC;

SELECT id, ipv4, port
FROM proxies
WHERE working = true
ORDER BY lastTested;

SELECT id
FROM proxies
WHERE ipv4 = ''
  AND port = 0;

INSERT INTO proxies (ipv4, port, lastTested, working)
VALUES ('$1', 0, 1609018536968587000, true);
