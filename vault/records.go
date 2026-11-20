package vault

type Params struct {
	BaseEnabled        bool `json:"base_enabled"`
	TraderswiftEnabled bool `json:"traderswift_enabled"`
}

type AccountLink struct {
	Nonce      uint64 `json:"nonce"`
	Owner      string `json:"owner"`
	HubAccount string `json:"hub_account"`
}

type StorageChoice struct {
	Owner     string      `json:"owner"`
	Locations []*Location `json:"locations"`
}

type CommitmentRoot struct {
	Owner string `json:"owner"`
	Day   string `json:"day"`
	Root  []byte `json:"root"`
}

type Grant struct {
	Owner             string `json:"owner"`
	GranteeCommitment []byte `json:"grantee_commitment"`
	Scope             string `json:"scope"`
	Expiry            uint64 `json:"expiry"`
	Revoked           bool   `json:"revoked"`
}

type Frontend struct {
	ProviderID    string `json:"provider_id"`
	Kind          string `json:"kind"`
	ServiceSigner string `json:"service_signer"`
	ServicePubkey []byte `json:"service_pubkey"`
	Endpoint      string `json:"endpoint"`
	CapacityClass string `json:"capacity_class"`
	CapacityBytes uint64 `json:"capacity_bytes"`
	Status        string `json:"status"`
}

type Provider struct {
	ProviderID              string   `json:"provider_id"`
	Owner                   string   `json:"owner"`
	ConsensusPubkey         []byte   `json:"consensus_pubkey"`
	Power                   int64    `json:"power"`
	Founding                bool     `json:"founding"`
	Vacant                  bool     `json:"vacant,omitempty"`
	RetiredConsensusPubkeys [][]byte `json:"retired_consensus_pubkeys,omitempty"`
}

type State struct {
	Providers         []*Provider       `json:"providers"`
	LastAdmissionTime int64             `json:"last_admission_time"`
	SchemaVersion     uint64            `json:"schema_version"`
	Params            Params            `json:"params"`
	Links             []*AccountLink    `json:"links"`
	Choices           []*StorageChoice  `json:"choices"`
	Roots             []*CommitmentRoot `json:"roots"`
	Grants            []*Grant          `json:"grants"`
	Frontends         []*Frontend       `json:"frontends"`
	RetiredPubkeys    [][]byte          `json:"retired_pubkeys,omitempty"`
}

type WhereRecord struct {
	Link   *AccountLink   `json:"link"`
	Choice *StorageChoice `json:"choice"`
}
