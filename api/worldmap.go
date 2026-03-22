package api

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "modernc.org/sqlite"
)

var worldMapDB *sql.DB

// InitWorldMap opens (or creates) the SQLite world map database.
func InitWorldMap(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS blocks (
			x           INTEGER NOT NULL,
			y           INTEGER NOT NULL,
			z           INTEGER NOT NULL,
			name        TEXT,    -- NULL = air/empty (we looked and nothing was there)
			state       TEXT,    -- JSON: block state/properties from inspect()
			extra       TEXT,    -- JSON object keyed by handler name; merged across updates
			observed_at INTEGER NOT NULL,  -- epoch ms from turtle heartbeat
			observed_by INTEGER NOT NULL,  -- turtle computer ID
			PRIMARY KEY (x, y, z)
		);
		CREATE INDEX IF NOT EXISTS idx_blocks_name ON blocks (name);
	`)
	if err != nil {
		return err
	}

	worldMapDB = db
	log.Println("[WorldMap] Initialized:", dbPath)
	return nil
}

// blockObservation is one block entry derived from a turtle heartbeat.
type blockObservation struct {
	X, Y, Z    int
	Name       *string                // nil = air
	State      interface{}            // raw block state from inspect()
	Extra      map[string]interface{} // handler-provided data; merged with existing
	ObservedAt int64
	ObservedBy int
}

// upsertBlock writes a single block observation to the world map.
// Extra data is merged with any existing extra data for that coordinate.
func upsertBlock(obs blockObservation) error {
	if worldMapDB == nil {
		return nil
	}

	var stateStr *string
	if obs.State != nil {
		b, err := json.Marshal(obs.State)
		if err == nil {
			s := string(b)
			stateStr = &s
		}
	}

	// Merge extra: fetch existing, overlay new keys, write back.
	var extraStr *string
	if len(obs.Extra) > 0 {
		merged := make(map[string]interface{})
		var existing sql.NullString
		_ = worldMapDB.QueryRow(
			"SELECT extra FROM blocks WHERE x=? AND y=? AND z=?",
			obs.X, obs.Y, obs.Z,
		).Scan(&existing)
		if existing.Valid && existing.String != "" {
			_ = json.Unmarshal([]byte(existing.String), &merged)
		}
		for k, v := range obs.Extra {
			merged[k] = v
		}
		b, _ := json.Marshal(merged)
		s := string(b)
		extraStr = &s
	}

	_, err := worldMapDB.Exec(`
		INSERT INTO blocks (x, y, z, name, state, extra, observed_at, observed_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(x, y, z) DO UPDATE SET
			name        = excluded.name,
			state       = excluded.state,
			extra       = COALESCE(excluded.extra, extra),
			observed_at = excluded.observed_at,
			observed_by = excluded.observed_by
	`, obs.X, obs.Y, obs.Z, obs.Name, stateStr, extraStr, obs.ObservedAt, obs.ObservedBy)
	return err
}

// extractBlockName pulls the "name" string from raw sight data.
// Sight entries are either empty ({}) meaning air, or a block table with "name".
func extractBlockName(raw interface{}) *string {
	if raw == nil {
		return nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok || len(m) == 0 {
		return nil // empty {} = air
	}
	name, ok := m["name"].(string)
	if !ok || name == "" {
		return nil
	}
	return &name
}

// extractBlockState pulls the "state" sub-table from raw sight data.
func extractBlockState(raw interface{}) interface{} {
	if m, ok := raw.(map[string]interface{}); ok {
		return m["state"]
	}
	return nil
}

// sightWorldCoords returns the absolute world coordinates for each sight
// direction given the turtle's current position and facing.
//
// Facing: 0=north(-z), 1=east(+x), 2=south(+z), 3=west(-x)
func sightWorldCoords(pos struct {
	X, Y, Z, R int
}) map[string][3]int {
	x, y, z, r := pos.X, pos.Y, pos.Z, pos.R
	out := map[string][3]int{
		"up":   {x, y + 1, z},
		"down": {x, y - 1, z},
	}
	switch r {
	case 0: // north
		out["front"] = [3]int{x, y, z - 1}
		out["left"]  = [3]int{x - 1, y, z}
		out["back"]  = [3]int{x, y, z + 1}
		out["right"] = [3]int{x + 1, y, z}
	case 1: // east
		out["front"] = [3]int{x + 1, y, z}
		out["left"]  = [3]int{x, y, z - 1}
		out["back"]  = [3]int{x - 1, y, z}
		out["right"] = [3]int{x, y, z + 1}
	case 2: // south
		out["front"] = [3]int{x, y, z + 1}
		out["left"]  = [3]int{x + 1, y, z}
		out["back"]  = [3]int{x, y, z - 1}
		out["right"] = [3]int{x - 1, y, z}
	case 3: // west
		out["front"] = [3]int{x - 1, y, z}
		out["left"]  = [3]int{x, y, z + 1}
		out["back"]  = [3]int{x + 1, y, z}
		out["right"] = [3]int{x, y, z - 1}
	}
	return out
}

// RecordTurtleSight extracts block observations from a turtle's sight data
// and upserts them into the world map. Called on every heartbeat.
func RecordTurtleSight(t Turtle) {
	if worldMapDB == nil {
		return
	}

	coords := sightWorldCoords(struct{ X, Y, Z, R int }{
		t.Pos.X, t.Pos.Y, t.Pos.Z, t.Pos.R,
	})

	sightFields := map[string]interface{}{
		"up":    t.Sight.Up,
		"down":  t.Sight.Down,
		"front": t.Sight.Front,
		"left":  t.Sight.Left,
		"right": t.Sight.Right,
		"back":  t.Sight.Back,
	}

	for dir, raw := range sightFields {
		if raw == nil {
			continue // direction not present (inspectAll disabled)
		}
		c, ok := coords[dir]
		if !ok {
			continue
		}
		obs := blockObservation{
			X:          c[0],
			Y:          c[1],
			Z:          c[2],
			Name:       extractBlockName(raw),
			State:      extractBlockState(raw),
			ObservedAt: int64(t.HeartBeat),
			ObservedBy: t.ID,
		}
		if err := upsertBlock(obs); err != nil {
			log.Printf("[WorldMap] upsert error at %d,%d,%d: %v", c[0], c[1], c[2], err)
		}
	}
}

// --- Query helpers used by MCP tools ---

type BlockRecord struct {
	X          int                    `json:"x"`
	Y          int                    `json:"y"`
	Z          int                    `json:"z"`
	Name       *string                `json:"name"`
	State      interface{}            `json:"state,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
	ObservedAt int64                  `json:"observed_at"`
	ObservedBy int                    `json:"observed_by"`
}

func scanBlockRow(rows *sql.Rows) (BlockRecord, error) {
	var b BlockRecord
	var stateStr, extraStr sql.NullString
	err := rows.Scan(&b.X, &b.Y, &b.Z, &b.Name, &stateStr, &extraStr, &b.ObservedAt, &b.ObservedBy)
	if err != nil {
		return b, err
	}
	if stateStr.Valid {
		_ = json.Unmarshal([]byte(stateStr.String), &b.State)
	}
	if extraStr.Valid {
		_ = json.Unmarshal([]byte(extraStr.String), &b.Extra)
	}
	return b, nil
}

// GetBlock returns the known data for a single coordinate, or nil if unknown.
func GetBlock(x, y, z int) (*BlockRecord, error) {
	rows, err := worldMapDB.Query(
		"SELECT x,y,z,name,state,extra,observed_at,observed_by FROM blocks WHERE x=? AND y=? AND z=?",
		x, y, z,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		b, err := scanBlockRow(rows)
		return &b, err
	}
	return nil, nil
}

// FindBlock returns all known coordinates where name contains the search string.
// Results capped at limit (max 500).
func FindBlock(nameSearch string, limit int) ([]BlockRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	rows, err := worldMapDB.Query(
		"SELECT x,y,z,name,state,extra,observed_at,observed_by FROM blocks WHERE name LIKE ? LIMIT ?",
		"%"+nameSearch+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []BlockRecord
	for rows.Next() {
		b, err := scanBlockRow(rows)
		if err != nil {
			continue
		}
		results = append(results, b)
	}
	return results, nil
}

// GetRegion returns all known blocks within a bounding box. Capped at 1000.
func GetRegion(x1, y1, z1, x2, y2, z2 int) ([]BlockRecord, error) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if z1 > z2 {
		z1, z2 = z2, z1
	}
	rows, err := worldMapDB.Query(`
		SELECT x,y,z,name,state,extra,observed_at,observed_by FROM blocks
		WHERE x BETWEEN ? AND ? AND y BETWEEN ? AND ? AND z BETWEEN ? AND ?
		LIMIT 1000
	`, x1, x2, y1, y2, z1, z2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []BlockRecord
	for rows.Next() {
		b, err := scanBlockRow(rows)
		if err != nil {
			continue
		}
		results = append(results, b)
	}
	return results, nil
}
