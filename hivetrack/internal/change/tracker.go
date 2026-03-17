package change

// Type represents the type of change for a change tracker entry.
type Type int

const (
	Added   Type = iota
	Updated Type = iota
	Deleted Type = iota
)

// Entry holds a single queued change.
type Entry struct {
	itemType   int
	item       any
	changeType Type
}

func NewEntry(itemType int, item any, changeType Type) Entry {
	return Entry{itemType: itemType, item: item, changeType: changeType}
}

func (e Entry) GetItemType() int    { return e.itemType }
func (e Entry) GetItem() any        { return e.item }
func (e Entry) GetChangeType() Type { return e.changeType }

// Tracker collects change entries for batch processing by SaveChanges.
type Tracker struct {
	changes []Entry
}

func NewTracker() *Tracker {
	return &Tracker{}
}

func (t *Tracker) Add(entry Entry) {
	t.changes = append(t.changes, entry)
}

func (t *Tracker) GetChanges() []Entry {
	return t.changes
}

func (t *Tracker) Clear() {
	t.changes = t.changes[:0]
}
