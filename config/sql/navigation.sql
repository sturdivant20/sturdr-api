-- name: make_navigation_table
CREATE TABLE IF NOT EXISTS navigation (
  sequence INTEGER NOT NULL PRIMARY KEY,
  week INTEGER NOT NULL CHECK (week >= 0),
  tow REAL NOT NULL CHECK (tow BETWEEN 0 AND 604800),
  n_sat INTEGER NOT NULL,
  latitude REAL NOT NULL,
  longitude REAL NOT NULL,
  altitude REAL NOT NULL,
  vn REAL NOT NULL,
  ve REAL NOT NULL,
  vd REAL NOT NULL,
  roll REAL NOT NULL,
  pitch REAL NOT NULL,
  yaw REAL NOT NULL,
  pdop REAL NOT NULL,
  hdop REAL NOT NULL,
  vdop REAL NOT NULL
);

CREATE INDEX idx_gps_time 
ON navigation (week DESC, tow DESC);

-- name: create_navigation
INSERT INTO navigation (sequence, week, tow, n_sat, latitude, longitude, altitude, vn, ve, vd, roll, pitch, yaw, pdop, hdop, vdop)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);

-- name: read_latest_navigation
SELECT sequence, week, tow, n_sat, latitude, longitude, altitude, vn, ve, vd, roll, pitch, yaw, pdop, hdop, vdop 
FROM navigation
ORDER BY week DESC, tow DESC
LIMIT 1;

-- name: read_queried_navigation
SELECT sequence, week, tow, n_sat, latitude, longitude, altitude, vn, ve, vd, roll, pitch, yaw, pdop, hdop, vdop 
FROM navigation
WHERE (week > $1) OR (week = $1 AND tow >= $2)
ORDER BY week ASC, tow ASC;

-- name: update_navigation
UPDATE navigation
SET sequence = $1, week = $2, tow = $3, n_sat = $4, latitude = $5, longitude = $6, altitude = $7, vn = $8, ve = $9, vd = $10, roll = $11, pitch = $12, yaw = $13, pdop = $14, hdop = $15, vdop = $16
WHERE sequence = $1;

-- name: delete_navigation
DELETE FROM navigation
WHERE sequence = $1;