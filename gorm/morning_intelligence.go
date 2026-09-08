package happywakeygorm

import "time"

type Tenant struct {
	ID            string    `gorm:"column:id;primaryKey"`
	AccountKind   string    `gorm:"column:account_kind;not null"`
	DisplayName   string    `gorm:"column:display_name;not null"`
	SeatLimit     uint32    `gorm:"column:seat_limit;not null"`
	PolicyVersion string    `gorm:"column:policy_version;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
}

func (Tenant) TableName() string { return "happy_wakey_tenants" }

type TenantMembership struct {
	TenantID  string    `gorm:"column:tenant_id;primaryKey"`
	SubjectID string    `gorm:"column:subject_id;primaryKey"`
	Role      string    `gorm:"column:role;not null"`
	State     string    `gorm:"column:state;not null"`
	JoinedAt  time.Time `gorm:"column:joined_at;not null"`
}

func (TenantMembership) TableName() string { return "happy_wakey_tenant_memberships" }

type ConnectorConsent struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	TenantID         string     `gorm:"column:tenant_id;not null;index:happy_wakey_connector_consents_scope,priority:1;uniqueIndex:happy_wakey_connector_consent_identity,priority:1"`
	SubjectID        string     `gorm:"column:subject_id;not null;index:happy_wakey_connector_consents_scope,priority:2;uniqueIndex:happy_wakey_connector_consent_identity,priority:2"`
	Connector        string     `gorm:"column:connector;not null;uniqueIndex:happy_wakey_connector_consent_identity,priority:3"`
	State            string     `gorm:"column:state;not null;index:happy_wakey_connector_consents_scope,priority:3"`
	ScopesJSON       []byte     `gorm:"column:scopes;type:jsonb;not null"`
	SourceAccountRef string     `gorm:"column:source_account_ref;not null;uniqueIndex:happy_wakey_connector_consent_identity,priority:4"`
	CredentialRef    *string    `gorm:"column:credential_ref"`
	GrantedAt        *time.Time `gorm:"column:granted_at"`
	ExpiresAt        *time.Time `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	PolicyVersion    string     `gorm:"column:policy_version;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null;index:happy_wakey_connector_consents_scope,priority:4,sort:desc"`
}

func (ConnectorConsent) TableName() string { return "happy_wakey_connector_consents" }

type SourceItemCandidate struct {
	ID                    string     `gorm:"column:id;type:uuid;primaryKey"`
	TenantID              string     `gorm:"column:tenant_id;not null;index:happy_wakey_candidates_pending,priority:1;uniqueIndex:happy_wakey_candidate_source_identity,priority:1"`
	SubjectID             string     `gorm:"column:subject_id;not null;index:happy_wakey_candidates_pending,priority:2;uniqueIndex:happy_wakey_candidate_source_identity,priority:2"`
	ConsentID             string     `gorm:"column:consent_id;type:uuid;not null"`
	Provider              string     `gorm:"column:provider;not null;uniqueIndex:happy_wakey_candidate_source_identity,priority:3"`
	SourceItemRef         string     `gorm:"column:source_item_ref;not null;uniqueIndex:happy_wakey_candidate_source_identity,priority:4"`
	SenderClass           string     `gorm:"column:sender_class;not null"`
	ReceivedAt            time.Time  `gorm:"column:received_at;not null;index:happy_wakey_candidates_pending,priority:3,sort:desc"`
	ThreadRef             string     `gorm:"column:thread_ref;not null"`
	EncryptedContentRef   string     `gorm:"column:encrypted_content_ref;not null"`
	ContentSHA256         string     `gorm:"column:content_sha256;not null"`
	HasDirectReplyRequest bool       `gorm:"column:has_direct_reply_request;not null"`
	DueAt                 *time.Time `gorm:"column:due_at"`
	RetentionExpiresAt    time.Time  `gorm:"column:retention_expires_at;not null"`
	CreatedAt             time.Time  `gorm:"column:created_at;not null"`
}

func (SourceItemCandidate) TableName() string { return "happy_wakey_source_item_candidates" }

type UsefulnessDecision struct {
	ID               string    `gorm:"column:id;type:uuid;primaryKey"`
	TenantID         string    `gorm:"column:tenant_id;not null;uniqueIndex:happy_wakey_usefulness_idempotency,priority:1"`
	SubjectID        string    `gorm:"column:subject_id;not null;uniqueIndex:happy_wakey_usefulness_idempotency,priority:2"`
	SourceItemID     string    `gorm:"column:source_item_id;type:uuid;not null"`
	Model            string    `gorm:"column:model;not null"`
	PolicyVersion    string    `gorm:"column:policy_version;not null"`
	ContentSHA256    string    `gorm:"column:content_sha256;not null"`
	Score            float64   `gorm:"column:score;not null"`
	DesignatedUseful bool      `gorm:"column:designated_useful;not null"`
	ReasonsJSON      []byte    `gorm:"column:reasons;type:jsonb;not null"`
	Rationale        string    `gorm:"column:rationale;not null"`
	IdempotencyKey   string    `gorm:"column:idempotency_key;not null;uniqueIndex:happy_wakey_usefulness_idempotency,priority:3"`
	EvaluatedAt      time.Time `gorm:"column:evaluated_at;not null"`
}

func (UsefulnessDecision) TableName() string { return "happy_wakey_usefulness_decisions" }

type SafeDeepLink struct {
	ID                       string    `gorm:"column:id;type:uuid;primaryKey"`
	TenantID                 string    `gorm:"column:tenant_id;not null;uniqueIndex:happy_wakey_safe_link_identity,priority:1"`
	SubjectID                string    `gorm:"column:subject_id;not null;uniqueIndex:happy_wakey_safe_link_identity,priority:2"`
	SourceItemID             string    `gorm:"column:source_item_id;type:uuid;not null;uniqueIndex:happy_wakey_safe_link_identity,priority:3"`
	DecisionID               string    `gorm:"column:decision_id;type:uuid;not null;uniqueIndex:happy_wakey_safe_link_identity,priority:4"`
	Provider                 string    `gorm:"column:provider;not null"`
	TargetURL                string    `gorm:"column:target_url;not null"`
	ExpiresAt                time.Time `gorm:"column:expires_at;not null"`
	RequiresReauthentication bool      `gorm:"column:requires_reauthentication;not null"`
	FeedFallbackAllowed      bool      `gorm:"column:feed_fallback_allowed;not null"`
	CreatedAt                time.Time `gorm:"column:created_at;not null"`
}

func (SafeDeepLink) TableName() string { return "happy_wakey_safe_deep_links" }

type MorningBriefing struct {
	ID                   string    `gorm:"column:id;type:uuid;primaryKey"`
	TenantID             string    `gorm:"column:tenant_id;not null;uniqueIndex:happy_wakey_briefing_subject_day,priority:1"`
	SubjectID            string    `gorm:"column:subject_id;not null;uniqueIndex:happy_wakey_briefing_subject_day,priority:2"`
	LocalDate            time.Time `gorm:"column:local_date;type:date;not null;uniqueIndex:happy_wakey_briefing_subject_day,priority:3"`
	TimeZone             string    `gorm:"column:time_zone;not null"`
	Title                string    `gorm:"column:title;not null"`
	CardsJSON            []byte    `gorm:"column:cards;type:jsonb;not null"`
	CardCount            uint32    `gorm:"column:card_count;not null"`
	SuppressedItemCount  uint32    `gorm:"column:suppressed_item_count;not null"`
	AudioURL             *string   `gorm:"column:audio_url"`
	AudioDurationSeconds *uint32   `gorm:"column:audio_duration_seconds"`
	GeneratedAt          time.Time `gorm:"column:generated_at;not null"`
	ValidUntil           time.Time `gorm:"column:valid_until;not null"`
}

func (MorningBriefing) TableName() string { return "happy_wakey_morning_briefings" }

type Embedding struct {
	ID             string    `gorm:"column:id;type:uuid;primaryKey"`
	TenantID       string    `gorm:"column:tenant_id;not null;uniqueIndex:happy_wakey_embedding_idempotency,priority:1"`
	SubjectID      string    `gorm:"column:subject_id;not null;uniqueIndex:happy_wakey_embedding_idempotency,priority:2"`
	ResourceType   string    `gorm:"column:resource_type;not null"`
	ResourceID     string    `gorm:"column:resource_id;not null"`
	Model          string    `gorm:"column:model;not null"`
	ContentHash    string    `gorm:"column:content_hash;not null"`
	Dimensions     uint32    `gorm:"column:dimensions;not null"`
	Embedding      []float32 `gorm:"column:embedding;type:real[];not null"`
	RetentionClass string    `gorm:"column:retention_class;not null"`
	IdempotencyKey string    `gorm:"column:idempotency_key;not null;uniqueIndex:happy_wakey_embedding_idempotency,priority:3"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

func (Embedding) TableName() string { return "happy_wakey_embeddings" }

type CorrelationFinding struct {
	ID                 string    `gorm:"column:id;type:uuid;primaryKey"`
	TenantID           string    `gorm:"column:tenant_id;not null"`
	SubjectID          string    `gorm:"column:subject_id;not null"`
	MetricA            string    `gorm:"column:metric_a;not null"`
	MetricB            string    `gorm:"column:metric_b;not null"`
	Coefficient        float64   `gorm:"column:coefficient;not null"`
	PValue             float64   `gorm:"column:p_value;not null"`
	ConfidenceLow      float64   `gorm:"column:confidence_low;not null"`
	ConfidenceHigh     float64   `gorm:"column:confidence_high;not null"`
	SampleSize         uint32    `gorm:"column:sample_size;not null"`
	Correction         string    `gorm:"column:correction;not null"`
	CausalClaimAllowed bool      `gorm:"column:causal_claim_allowed;not null"`
	DiscoveredAt       time.Time `gorm:"column:discovered_at;not null"`
}

func (CorrelationFinding) TableName() string { return "happy_wakey_correlation_findings" }

type ChatSession struct {
	ID                  string     `gorm:"column:id;type:uuid;primaryKey"`
	TenantID            string     `gorm:"column:tenant_id;not null"`
	SubjectID           string     `gorm:"column:subject_id;not null"`
	Audience            string     `gorm:"column:audience;not null"`
	AllowedSearchScopes []byte     `gorm:"column:allowed_search_scopes;type:jsonb;not null"`
	OresChatSessionRef  string     `gorm:"column:ores_chat_session_ref;not null"`
	OpenedAt            time.Time  `gorm:"column:opened_at;not null"`
	ExpiresAt           time.Time  `gorm:"column:expires_at;not null"`
	ClosedAt            *time.Time `gorm:"column:closed_at"`
}

func (ChatSession) TableName() string { return "happy_wakey_chat_sessions" }
