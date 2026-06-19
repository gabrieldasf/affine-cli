package canvaswrite

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
)

type TransformOptions struct {
	DocID        string
	IDs          []string
	Move         string
	Resize       string
	Align        string
	Distribute   string
	DisplayMode  string
	Metadata     []string
	BackupTarget string
}

type TransformPlan struct {
	PlanType     string               `json:"plan_type"`
	PlanID       string               `json:"plan_id"`
	DryRun       bool                 `json:"dry_run"`
	DocID        string               `json:"doc_id,omitempty"`
	Source       DiffSource           `json:"source"`
	AffectedIDs  []string             `json:"affected_ids"`
	Operations   []TransformOperation `json:"operations"`
	Integrity    DocIntegrityResult   `json:"integrity"`
	BackupTarget string               `json:"backup_target,omitempty"`
	Rollback     TransformProof       `json:"rollback"`
	Proof        TransformProof       `json:"proof"`
	Warnings     []string             `json:"warnings,omitempty"`
}

type TransformOperation struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type TransformProof struct {
	Required bool     `json:"required"`
	Fields   []string `json:"fields"`
}

func BuildTransformPlan(selectors SearchResult, opts TransformOptions) (TransformPlan, error) {
	entities := selectedEntities(selectors.Entities, opts.IDs)
	if len(entities) == 0 {
		return TransformPlan{}, fmt.Errorf("no selected entities; pass selector JSON or --id")
	}
	for _, entity := range entities {
		if len(entity.IntegrityIssues) > 0 {
			return TransformPlan{}, fmt.Errorf("selector %q has integrity issues: %s", entity.ID, strings.Join(entity.IntegrityIssues, ", "))
		}
	}
	metadata, err := parseMetadata(opts.Metadata)
	if err != nil {
		return TransformPlan{}, err
	}
	var ops []TransformOperation
	if opts.Move != "" {
		delta, err := parsePair(opts.Move, "--move")
		if err != nil {
			return TransformPlan{}, err
		}
		for _, e := range entities {
			op, err := geometryOperation("move", e, func(xywh []float64) []float64 {
				return []float64{xywh[0] + delta[0], xywh[1] + delta[1], xywh[2], xywh[3]}
			})
			if err != nil {
				return TransformPlan{}, err
			}
			ops = append(ops, op)
		}
	}
	if opts.Resize != "" {
		size, err := parsePair(opts.Resize, "--resize")
		if err != nil {
			return TransformPlan{}, err
		}
		if size[0] <= 0 || size[1] <= 0 {
			return TransformPlan{}, fmt.Errorf("--resize width and height must be positive")
		}
		for _, e := range entities {
			op, err := geometryOperation("resize", e, func(xywh []float64) []float64 {
				return []float64{xywh[0], xywh[1], size[0], size[1]}
			})
			if err != nil {
				return TransformPlan{}, err
			}
			ops = append(ops, op)
		}
	}
	if opts.Align != "" {
		aligned, err := alignOperations(entities, opts.Align)
		if err != nil {
			return TransformPlan{}, err
		}
		ops = append(ops, aligned...)
	}
	if opts.Distribute != "" {
		distributed, err := distributeOperations(entities, opts.Distribute)
		if err != nil {
			return TransformPlan{}, err
		}
		ops = append(ops, distributed...)
	}
	if opts.DisplayMode != "" {
		for _, e := range entities {
			ops = append(ops, TransformOperation{Kind: "set_display_mode", ID: e.ID, Before: e.DisplayMode, After: opts.DisplayMode})
		}
	}
	for key, value := range metadata {
		for _, e := range entities {
			ops = append(ops, TransformOperation{Kind: "set_metadata", ID: e.ID, Before: map[string]string{key: ""}, After: map[string]string{key: value}})
		}
	}
	if len(ops) == 0 {
		return TransformPlan{}, fmt.Errorf("no transform operation requested")
	}
	affected := affectedIDs(ops)
	integrity := DocIntegrityResult{DocID: opts.DocID, OK: true, Summary: map[string]int{}}
	plan := TransformPlan{
		PlanType:     "canvas_transform",
		DryRun:       true,
		DocID:        opts.DocID,
		Source:       DiffSource{Mode: selectors.SourceMode, Timestamp: selectors.Timestamp, Count: selectors.Count},
		AffectedIDs:  affected,
		Operations:   ops,
		Integrity:    integrity,
		BackupTarget: opts.BackupTarget,
		Rollback:     TransformProof{Required: true, Fields: []string{"before_snapshot", "delta", "affected_ids"}},
		Proof:        TransformProof{Required: true, Fields: []string{"pre_integrity", "post_integrity", "reload_verification"}},
	}
	plan.PlanID = transformPlanID(plan)
	return plan, nil
}

func selectedEntities(entities []SearchEntity, ids []string) []SearchEntity {
	if len(ids) == 0 {
		return entities
	}
	byID := searchEntitiesByID(entities)
	var out []SearchEntity
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if e, ok := byID[id]; ok {
			out = append(out, e)
		} else {
			out = append(out, SearchEntity{ID: id, Kind: "block"})
		}
	}
	return out
}

func geometryOperation(kind string, e SearchEntity, transform func([]float64) []float64) (TransformOperation, error) {
	if len(e.XYWH) != 4 {
		return TransformOperation{}, fmt.Errorf("%s requires selector %q to include xywh", kind, e.ID)
	}
	after := transform(append([]float64(nil), e.XYWH...))
	if after[2] <= 0 || after[3] <= 0 {
		return TransformOperation{}, fmt.Errorf("%s creates invalid geometry for %q", kind, e.ID)
	}
	return TransformOperation{Kind: kind, ID: e.ID, Before: e.XYWH, After: after}, nil
}

func alignOperations(entities []SearchEntity, mode string) ([]TransformOperation, error) {
	if len(entities) < 2 {
		return nil, fmt.Errorf("--align requires at least two selected entities")
	}
	base, err := alignTarget(entities, mode)
	if err != nil {
		return nil, err
	}
	var ops []TransformOperation
	for _, e := range entities {
		op, err := geometryOperation("align_"+mode, e, func(xywh []float64) []float64 {
			switch mode {
			case "left":
				xywh[0] = base
			case "right":
				xywh[0] = base - xywh[2]
			case "top":
				xywh[1] = base
			case "bottom":
				xywh[1] = base - xywh[3]
			case "center-x":
				xywh[0] = base - xywh[2]/2
			case "center-y":
				xywh[1] = base - xywh[3]/2
			}
			return xywh
		})
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func alignTarget(entities []SearchEntity, mode string) (float64, error) {
	first := entities[0].XYWH
	if len(first) != 4 {
		return 0, fmt.Errorf("--align requires selector %q to include xywh", entities[0].ID)
	}
	switch mode {
	case "left":
		return first[0], nil
	case "right":
		return first[0] + first[2], nil
	case "top":
		return first[1], nil
	case "bottom":
		return first[1] + first[3], nil
	case "center-x":
		return first[0] + first[2]/2, nil
	case "center-y":
		return first[1] + first[3]/2, nil
	default:
		return 0, fmt.Errorf("--align must be left, right, top, bottom, center-x, or center-y")
	}
}

func distributeOperations(entities []SearchEntity, axis string) ([]TransformOperation, error) {
	if len(entities) < 3 {
		return nil, fmt.Errorf("--distribute requires at least three selected entities")
	}
	if axis != "horizontal" && axis != "vertical" {
		return nil, fmt.Errorf("--distribute must be horizontal or vertical")
	}
	sorted := append([]SearchEntity(nil), entities...)
	for _, e := range sorted {
		if len(e.XYWH) != 4 {
			return nil, fmt.Errorf("--distribute requires selector %q to include xywh", e.ID)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		if axis == "horizontal" {
			return sorted[i].XYWH[0] < sorted[j].XYWH[0]
		}
		return sorted[i].XYWH[1] < sorted[j].XYWH[1]
	})
	start := sorted[0].XYWH[0]
	end := sorted[len(sorted)-1].XYWH[0]
	if axis == "vertical" {
		start = sorted[0].XYWH[1]
		end = sorted[len(sorted)-1].XYWH[1]
	}
	step := (end - start) / float64(len(sorted)-1)
	var ops []TransformOperation
	for i, e := range sorted {
		target := start + float64(i)*step
		op, err := geometryOperation("distribute_"+axis, e, func(xywh []float64) []float64 {
			if axis == "horizontal" {
				xywh[0] = target
			} else {
				xywh[1] = target
			}
			return xywh
		})
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func parsePair(raw, flag string) ([]float64, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("%s must be two comma-separated numbers", flag)
	}
	out := make([]float64, 2)
	for i, part := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be two comma-separated numbers", flag)
		}
		out[i] = v
	}
	return out, nil
}

func parseMetadata(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, raw := range values {
		key, value, ok := strings.Cut(raw, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("--set must use key=value")
		}
		out[key] = strings.TrimSpace(value)
	}
	return out, nil
}

func affectedIDs(ops []TransformOperation) []string {
	seen := map[string]bool{}
	for _, op := range ops {
		seen[op.ID] = true
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func transformPlanID(plan TransformPlan) string {
	h := fnv.New32a()
	for _, id := range plan.AffectedIDs {
		_, _ = h.Write([]byte(id))
	}
	for _, op := range plan.Operations {
		_, _ = h.Write([]byte(op.Kind + op.ID + fmt.Sprint(op.After)))
	}
	return fmt.Sprintf("canvas-transform-%08x", h.Sum32())
}
