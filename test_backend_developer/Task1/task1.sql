WITH SpotDomains AS (
  SELECT
    name,
    website,
    COALESCE(
      SUBSTRING(
        website FROM '^(?:https?://)?(?:www\.)?([^/]+)'
      ),
      ''
    ) AS domain
  FROM
    "MY_TABLE"
)
SELECT
  name AS "spot name",
  domain,
  COUNT(*) AS "count number for domain"
FROM
  SpotDomains
WHERE
  domain <> ''
GROUP BY
  name,
  domain
HAVING
  COUNT(*) > 1
ORDER BY
  COUNT(*) DESC;


-- For this exercice I used Docker to create a container called database and populated that with the spots.sql data

-- This is the command I used:
-- docker run -d --name database -e POSTGRES_PASSWORD=***** -p 5432:5432 postgis/postgis

-- Populate the database with the spots data:
-- psql -h 0.0.0.0 -d postgres -U postgres < spots.sql

-- Connect to the database via the psql CLI:
-- psql -h 0.0.0.0 -d postgres -U postgres