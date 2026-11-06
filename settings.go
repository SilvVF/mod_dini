package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"sync"

	"go.etcd.io/bbolt"
)

// Category is the left-hand navigation entry a setting belongs to.
type Category string

const (
	CategoryGeneral    Category = "general"
	CategoryAppearance Category = "appearance"
	CategoryMods       Category = "mods"
	CategoryInGame     Category = "in_game"
	CategoryGameBanana Category = "gamebanana"
	CategoryFolders    Category = "folders"
	CategoryTools      Category = "tools"
	CategoryAdvanced   Category = "advanced"
	CategoryAbout      Category = "about"
)

// Categories lists every category in navigation order.
var Categories = []Category{
	CategoryGeneral, CategoryAppearance, CategoryMods, CategoryInGame,
	CategoryGameBanana, CategoryFolders, CategoryTools, CategoryAdvanced, CategoryAbout,
}

// Kind tells the UI which control to render, and fixes the JSON shape of the value.
type Kind string

const (
	KindToggle      Kind = "toggle"      // value: bool
	KindDropdown    Kind = "dropdown"    // value: string (one of Options)
	KindSegmented   Kind = "segmented"   // value: string (one of Options), e.g. Ask | Tray | Quit
	KindMultiSelect Kind = "multiselect" // value: []string (subset of Options)
	KindText        Kind = "text"        // value: string
	KindNumber      Kind = "number"      // value: float64, bounded by Min/Max
)

// Option is one choice of a dropdown, segmented control or multiselect.
type Option struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type SettingChange struct {
	Kind  Kind            `json:"kind"`
	Value json.RawMessage `json:"value"`
}

// Action is an extra button shown on a row (e.g. "Customize..." next to Theme).
// The UI reports the click back using ID; it carries no value.
type Action struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Source string `json:"-"`
}

// Setting is one row in the settings screen.
type Setting struct {
	// Key is the stable unique identifier, e.g. "general.language". It is the bolt key.
	Key         string   `json:"key"`
	Category    Category `json:"category"`
	Section     string   `json:"section,omitempty"` // optional heading inside a category, e.g. "Appearance"
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"` // icon name resolved by the frontend
	Kind        Kind     `json:"kind"`

	Editable bool `json:"editable"`

	Options []Option `json:"options,omitempty"` // dropdown / segmented / multiselect
	Min     *float64 `json:"min,omitempty"`     // number
	Max     *float64 `json:"max,omitempty"`     // number
	Step    *float64 `json:"step,omitempty"`    // number
	Actions []Action `json:"actions,omitempty"`

	// Keywords feed the "Search settings..." box in addition to title/description.
	Keywords []string `json:"keywords,omitempty"`
	// Order is display order; Store.Sync assigns it from the definition order.
	Order int `json:"order"`

	// Default and Value hold JSON matching Kind. Value is empty until the user changes it.
	Default json.RawMessage `json:"default"`
	Value   json.RawMessage `json:"value,omitempty"`
	Source  string          `json:"-"`
}

var SETTINGS_BUCKET = []byte("settings")

// ErrSettingsNotFound is returned when a key has no setting.
var ErrSettingsNotFound = errors.New("setting not found")

// SettingsStore persists Settings as JSON in a bolt bucket, keyed by Setting.Key.
type SettingsStore struct {
	db *bbolt.DB

	// mu serialises writes so changes are emitted in the order they were committed.
	mu sync.Mutex

	// changes receives the updated Setting after every write that changed its
	// effective value. The small buffer only absorbs bursts while the listener
	// is busy; when it is full (no listener) new events are dropped.
	changes       chan Setting
	actions       map[string]func() error
	actionSources map[string]string
}

func NewStore(db *bbolt.DB) (*SettingsStore, error) {
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(SETTINGS_BUCKET)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &SettingsStore{
		db:            db,
		changes:       make(chan Setting, 32),
		actions:       make(map[string]func() error),
		actionSources: make(map[string]string),
	}, nil
}

// Changes returns the channel of changed settings (from SetValue and Reset).
// Typical use with Wails:
//
//	go func() {
//		for s := range store.Changes() {
//			runtime.EventsEmit(ctx, "settings:changed", s)
//		}
//	}()
func (st *SettingsStore) Changes() <-chan Setting { return st.changes }

func checkUniqueKeys(defs []Setting) {
	seen := make(map[string]Setting, len(defs))

	for _, s := range defs {
		if s.Key == "" {
			panic(fmt.Sprintf("setting with empty key defined at %s", s.Source))
		}
		if prev, ok := seen[s.Key]; ok {
			panic(fmt.Sprintf(
				"duplicate setting key %q\n  first defined at %s\n  defined again at %s",
				s.Key, prev.Source, s.Source,
			))
		}
		seen[s.Key] = s
	}
}

func (st *SettingsStore) checkActions(defs []Setting) {
	for _, s := range defs {
		seen := make(map[string]string, len(s.Actions)) // action id -> source

		for _, a := range s.Actions {
			if prevSrc, dup := seen[a.ID]; dup {
				panic(fmt.Sprintf(
					"duplicate action %q on setting %q\n  first defined at %s\n  defined again at %s",
					a.ID, s.Key, prevSrc, a.Source,
				))
			}
			seen[a.ID] = a.Source

			key := fmt.Sprintf("%s:%s", s.Key, a.ID)
			if _, ok := st.actions[key]; !ok {
				slog.Warn(
					fmt.Sprintf("action handler not register for %s %s \n\t handlerKey %s \n\t defined at %s", s.Key, a.ID, key, a.Source),
				)
			}
		}
	}
}

// Sync reconciles the stored settings with the definitions shipped in code.
// Call it once at startup. The definition wins for everything (title, options,
// order, ...) except the user's Value, which is kept when it is still valid.
// Settings that no longer exist in defs are removed. It does not emit changes;
// the UI loads the initial state with All().
func (st *SettingsStore) Sync(defs []Setting) error {
	checkUniqueKeys(defs)
	st.checkActions(defs)

	return st.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SETTINGS_BUCKET)
		live := make(map[string]struct{}, len(defs))

		for i, def := range defs {
			if err := def.validate(); err != nil {
				return fmt.Errorf("setting %q: %w", def.Key, err)
			}
			def.Order = i
			def.Value = nil

			var old Setting
			if raw := b.Get([]byte(def.Key)); raw != nil && json.Unmarshal(raw, &old) == nil {
				if len(old.Value) > 0 && def.validateValue(old.Value) == nil {
					def.Value = old.Value
				}
			}

			if err := put(b, def); err != nil {
				return err
			}
			live[def.Key] = struct{}{}
		}

		var stale [][]byte
		_ = b.ForEach(func(k, _ []byte) error {
			if _, ok := live[string(k)]; !ok {
				stale = append(stale, append([]byte(nil), k...)) // k is only valid inside the tx callback
			}
			return nil
		})
		for _, k := range stale {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// All returns every setting in display order.
func (st *SettingsStore) All() ([]Setting, error) {
	var out []Setting
	err := st.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(SETTINGS_BUCKET).ForEach(func(_, v []byte) error {
			var s Setting
			if err := json.Unmarshal(v, &s); err != nil {
				return err
			}
			out = append(out, s)
			return nil
		})
	})
	sort.SliceStable(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out, err
}

func GetSetting[T any](st *SettingsStore, key string) (T, error) {
	var value T
	s, err := st.Get(key)
	if err != nil {
		return value, err
	}
	return Get[T](s)
}

func (st *SettingsStore) Get(key string) (Setting, error) {
	var s Setting
	err := st.db.View(func(tx *bbolt.Tx) error {
		raw := tx.Bucket(SETTINGS_BUCKET).Get([]byte(key))
		if raw == nil {
			return ErrSettingsNotFound
		}
		return json.Unmarshal(raw, &s)
	})
	return s, err
}

// Reset clears the user's value so the default applies again.
func (st *SettingsStore) Reset(key string) (Setting, error) {
	return st.update(key, func(s *Setting) error { s.Value = nil; return nil })
}

func (st *SettingsStore) update(key string, fn func(*Setting) error) (Setting, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	var (
		s       Setting
		changed bool
	)
	err := st.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SETTINGS_BUCKET)
		raw := b.Get([]byte(key))
		if raw == nil {
			return ErrSettingsNotFound
		}
		s = Setting{}
		if err := json.Unmarshal(raw, &s); err != nil {
			return err
		}
		before := s.Current()
		if err := fn(&s); err != nil {
			return err
		}
		changed = !bytes.Equal(before, s.Current())
		return put(b, s)
	})
	if err != nil {
		return s, err
	}
	if changed {
		// Only after the commit succeeded. Never blocks: if nobody is
		// listening (buffer full) the event is simply not sent.
		select {
		case st.changes <- s:
		default:
		}
	}
	return s, nil
}

func put(b *bbolt.Bucket, s Setting) error {
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return b.Put([]byte(s.Key), raw)
}

// Current returns the user's value, falling back to the default.
func (s Setting) Current() json.RawMessage {
	if len(s.Value) > 0 {
		return s.Value
	}
	return s.Default
}

func (st *SettingsStore) DoAction(skey string, action string) error {
	if action, ok := st.actions[fmt.Sprintf("%s:%s", skey, action)]; ok && action != nil {
		return action()
	}
	return fmt.Errorf("actions not registerd in store for setting: %s, action: %s", skey, action)
}

// Get decodes the current value, e.g. Get[bool](s) or Get[[]string](s).
func Get[T any](s Setting) (T, error) {
	var v T
	err := json.Unmarshal(s.Current(), &v)
	return v, err
}

func (s Setting) hasOption(v string) bool {
	return slices.ContainsFunc(s.Options, func(o Option) bool { return o.Value == v })
}

func SetValueOf[T any](st *SettingsStore, key string, v T) (Setting, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return Setting{}, err
	}

	return st.SetValue(key, raw)
}

func (st *SettingsStore) SetValue(key string, change json.RawMessage) (Setting, error) {
	return st.update(key, func(s *Setting) error {

		var compact bytes.Buffer
		if err := json.Compact(&compact, change); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		value := json.RawMessage(compact.Bytes())

		err := s.validateValue(value)
		if err != nil {
			return err
		}

		s.Value = value

		return nil
	})
}

// ValidateValue checks that raw has the right shape for the setting's Kind.
func (s Setting) validateValue(raw json.RawMessage) error {
	switch s.Kind {
	case KindToggle:
		var v bool
		return json.Unmarshal(raw, &v)
	case KindText:
		var v string
		return json.Unmarshal(raw, &v)
	case KindDropdown, KindSegmented:
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if !s.hasOption(v) {
			return fmt.Errorf("%q is not one of the options", v)
		}
	case KindMultiSelect:
		var vs []string
		if err := json.Unmarshal(raw, &vs); err != nil {
			return err
		}
		for _, v := range vs {
			if !s.hasOption(v) {
				return fmt.Errorf("%q is not one of the options", v)
			}
		}
	case KindNumber:
		var v float64
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if s.Min != nil && v < *s.Min {
			return fmt.Errorf("%v is below the minimum %v", v, *s.Min)
		}
		if s.Max != nil && v > *s.Max {
			return fmt.Errorf("%v is above the maximum %v", v, *s.Max)
		}
	default:
		return fmt.Errorf("unknown kind %q", s.Kind)
	}
	return nil
}

func (s Setting) validate() error {
	switch {
	case s.Key == "":
		return errors.New("empty key")
	case s.Title == "":
		return errors.New("empty title")
	case !slices.Contains(Categories, s.Category):
		return fmt.Errorf("unknown category %q", s.Category)
	}
	switch s.Kind {
	case KindDropdown, KindSegmented, KindMultiSelect:
		if len(s.Options) == 0 {
			return errors.New("options required")
		}
	}
	if err := s.validateValue(s.Default); err != nil {
		return fmt.Errorf("default: %w", err)
	}
	return nil
}

// ---- constructors -----------------------------------------------------------

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func callerLocation(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func (st *SettingsStore) RegisterActionHandler(settingKey, actionKey string, handler func() error) {
	st.registerActionHandler(settingKey, actionKey, handler, callerLocation(1))
}

func (st *SettingsStore) registerActionHandler(settingKey, actionKey string, handler func() error, source string) {
	key := fmt.Sprintf("%s:%s", settingKey, actionKey)

	if prev, ok := st.actionSources[key]; ok {
		panic(fmt.Sprintf(
			"duplicate action handler %q\n  first registered at %s\n  registered again at %s",
			key, prev, source,
		))
	}

	st.actions[key] = handler
	st.actionSources[key] = source
}

func (s Setting) WithAction(id, label string) Setting {
	s.Actions = append(s.Actions, Action{ID: id, Label: label, Source: callerLocation(1)})
	return s
}

func (s Setting) WithActionHandler(store *SettingsStore, id, label string, handler func() error) Setting {
	src := callerLocation(1) // the line in DefaultSettings
	store.registerActionHandler(s.Key, id, handler, src)
	s.Actions = append(s.Actions, Action{ID: id, Label: label, Source: src})
	return s
}

func newSetting(key string, c Category, kind Kind, title, desc string, def any) Setting {
	return Setting{
		Key:         key,
		Category:    c,
		Kind:        kind,
		Title:       title,
		Description: desc,
		Editable:    true,
		Default:     mustJSON(def),
		Source:      callerLocation(2),
	}
}

func NewToggle(key string, c Category, title, desc string, def bool) Setting {
	return newSetting(key, c, KindToggle, title, desc, def)
}

func NewDropdown(key string, c Category, title, desc string, opts []Option, def string) Setting {
	s := newSetting(key, c, KindDropdown, title, desc, def)
	s.Options = opts
	return s
}

func NewSegmented(key string, c Category, title, desc string, opts []Option, def string) Setting {
	s := newSetting(key, c, KindSegmented, title, desc, def)
	s.Options = opts
	return s
}

func NewMultiSelect(key string, c Category, title, desc string, opts []Option, def []string) Setting {
	if def == nil {
		def = []string{}
	}
	s := newSetting(key, c, KindMultiSelect, title, desc, def)
	s.Options = opts
	return s
}

func NewText(key string, c Category, title, desc, def string) Setting {
	return newSetting(key, c, KindText, title, desc, def)
}

func NewNumber(key string, c Category, title, desc string, def, min, max, step float64) Setting {
	s := newSetting(key, c, KindNumber, title, desc, def)
	s.Min, s.Max, s.Step = &min, &max, &step
	return s
}

// ---- builder-style modifiers ---------------------------------------------------

func (s Setting) WithSection(name string) Setting  { s.Section = name; return s }
func (s Setting) WithIcon(icon string) Setting     { s.Icon = icon; return s }
func (s Setting) WithKeywords(k ...string) Setting { s.Keywords = k; return s }
func (s Setting) MakeUneditable() Setting          { s.Editable = false; return s }
