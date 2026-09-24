package api

// Typed ProductPlan resource IDs. Endpoint methods take these instead of
// strings, and a request path is built only by route.with, which accepts
// nothing but pathID values. So an ID reaches a URL only after its segment
// method has validated and escaped it (safeSeg), under the field name the
// caller used: an unvalidated ID cannot be interpolated by construction.
//
// The tools layer converts decoded string arguments at the boundary, for
// example api.BarID(a.BarID).

// pathID is implemented by every typed resource ID.
type pathID interface {
	// segment returns the validated, escaped path segment, or a
	// validation error naming the ID's field.
	segment() (string, error)
}

type (
	// RoadmapID identifies a roadmap.
	RoadmapID string
	// LaneID identifies a lane on a roadmap.
	LaneID string
	// MilestoneID identifies a milestone on a roadmap.
	MilestoneID string
	// BarID identifies a bar (roadmap item).
	BarID string
	// ConnectionID identifies a dependency connection from a bar.
	ConnectionID string
	// LinkID identifies an external link on a bar.
	LinkID string
	// ObjectiveID identifies an objective (OKR).
	ObjectiveID string
	// KeyResultID identifies a key result under an objective.
	KeyResultID string
	// IdeaID identifies a discovery idea.
	IdeaID string
	// OpportunityID identifies a discovery opportunity.
	OpportunityID string
	// IdeaFormID identifies an idea form.
	IdeaFormID string
	// LaunchID identifies a launch.
	LaunchID string
	// SectionID identifies a checklist section in a launch.
	SectionID string
	// TaskID identifies a task in a launch.
	TaskID string
)

func (id RoadmapID) segment() (string, error) { return safeSeg("roadmap_id", string(id)) }

func (id LaneID) segment() (string, error) { return safeSeg("lane_id", string(id)) }

func (id MilestoneID) segment() (string, error) { return safeSeg("milestone_id", string(id)) }

func (id BarID) segment() (string, error) { return safeSeg("bar_id", string(id)) }

func (id ConnectionID) segment() (string, error) { return safeSeg("connection_id", string(id)) }

func (id LinkID) segment() (string, error) { return safeSeg("link_id", string(id)) }

func (id ObjectiveID) segment() (string, error) { return safeSeg("objective_id", string(id)) }

func (id KeyResultID) segment() (string, error) { return safeSeg("key_result_id", string(id)) }

func (id IdeaID) segment() (string, error) { return safeSeg("idea_id", string(id)) }

func (id OpportunityID) segment() (string, error) { return safeSeg("opportunity_id", string(id)) }

func (id IdeaFormID) segment() (string, error) { return safeSeg("idea_form_id", string(id)) }

func (id LaunchID) segment() (string, error) { return safeSeg("launch_id", string(id)) }

func (id SectionID) segment() (string, error) { return safeSeg("section_id", string(id)) }

func (id TaskID) segment() (string, error) { return safeSeg("task_id", string(id)) }
