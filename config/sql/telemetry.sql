-- name: create_telemetry
INSERT INTO navigation (sequence, week, tow, n_sat, latitude, longitude, altitude, vn, ve, vd, roll, pitch, yaw, pdop, hdop, vdop)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);

INSERT INTO satellites (sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23);

-- name: read_telemetry
SELECT sequence, week, tow, n_sat, latitude, longitude, altitude, vn, ve, vd, roll, pitch, yaw, pdop, hdop, vdop 
FROM navigation
ORDER BY week DESC, tow DESC
LIMIT 1;

SELECT sequence, week, tow, n_sat, latitude, longitude, altitude, vn, ve, vd, roll, pitch, yaw, pdop, hdop, vdop 
FROM navigation
WHERE (week > $1) OR (week = $1 AND tow >= $2)
ORDER BY week ASC, tow ASC;

SELECT sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql 
FROM satellites 
WHERE (week, tow) = (SELECT week, tow FROM satellites ORDER BY week DESC, tow DESC LIMIT 1)
ORDER BY prn ASC;

SELECT sequence, week, tow, prn, health, x, y, z, vx, vy, vz, doppler, psr, adr, azimuth, elevation, cno, ie, ip, il, qe, qp, ql 
FROM satellites 
WHERE (week > $1) OR (week = $1 AND tow >= $2)
ORDER BY week ASC, tow ASC;

-- name: update_telemetry
UPDATE navigation
SET sequence = $1, week = $2, tow = $3, n_sat = $4, latitude = $5, longitude = $6, altitude = $7, vn = $8, ve = $9, vd = $10, roll = $11, pitch = $12, yaw = $13, pdop = $14, hdop = $15, vdop = $16
WHERE sequence = $1;

UPDATE satellites
SET sequence = $1, week = $2, tow = $3, prn = $4, health = $5, x = $6, y = $7, z = $8, vx = $9, vy = $10, vz = $11, doppler = $12, psr = $13, adr = $14, azimuth = $15, elevation = $16, cno = $17, ie = $18, ip = $19, il = $20, qe = $21, qp = $22, ql = $23
WHERE row = $24;

-- name: delete_telemetry
DELETE FROM navigation
WHERE sequence = $1;

DELETE FROM satellites
WHERE row = $1;