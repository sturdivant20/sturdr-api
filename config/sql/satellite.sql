-- name: make_satellite_table
CREATE TABLE IF NOT EXISTS satellites (
  row INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  sequence INTEGER NOT NULL CHECK (sequence >= 0),
  week INTEGER NOT NULL CHECK (week >= 0),
  tow REAL NOT NULL CHECK (tow BETWEEN 0 AND 604800),
  prn INTEGER NOT NULL CHECK (prn >= 0),
  health INTEGER NOT NULL,
  x REAL NOT NULL,
  y REAL NOT NULL,
  z REAL NOT NULL,
  vx REAL NOT NULL,
  vy REAL NOT NULL,
  vz REAL NOT NULL,
  doppler REAL NOT NULL,
  psr REAL NOT NULL,
  adr REAL NOT NULL,
  azimuth REAL NOT NULL,
  elevation REAL NOT NULL,
  cno REAL NOT NULL,
  ie REAL NOT NULL,
  ip REAL NOT NULL,
  il REAL NOT NULL,
  qe REAL NOT NULL,
  qp REAL NOT NULL,
  ql REAL NOT NULL,
  FOREIGN KEY(sequence) REFERENCES navigation(sequence) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_satellite_gps 
ON satellites (prn, week DESC, tow DESC);

-- name: create_satellite
INSERT INTO satellites (sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23);

-- name: read_latest_satellite
SELECT sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql 
FROM satellites 
WHERE (week, tow) = (SELECT week, tow FROM satellites ORDER BY week DESC, tow DESC LIMIT 1)
ORDER BY prn ASC;

-- name: read_queried_satellite
SELECT sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql 
FROM satellites 
WHERE (week > $1) OR (week = $1 AND tow >= $2)
ORDER BY week ASC, tow ASC;

-- name: read_latest_specific_satellite
SELECT sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql 
FROM satellites 
WHERE prn = $1 
ORDER BY week DESC, tow DESC LIMIT 1;

-- name: read_queried_specific_satellite
SELECT sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql 
FROM satellites 
WHERE ((week > $1) OR (week = $1 AND tow >= $2)) AND prn = $3
ORDER BY week ASC, tow ASC, prn ASC;

-- name: update_satellite
UPDATE satellites
SET sequence = $1, week = $2, tow = $3, prn = $4, health = $5, x = $6, y = $7, z = $8, vx = $9, vy = $10, vz = $11, doppler = $12, psr = $13, adr = $14, azimuth = $15, elevation = $16, cno = $17, ie = $18, ip = $19, il = $20, qe = $21, qp = $22, ql = $23
WHERE row = $24;

-- name: delete_satellite
DELETE FROM satellites
WHERE row = $1;