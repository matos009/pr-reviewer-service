package domain

type Team struct {
	Name    string
	Members []TeamMember
}

type TeamMember struct {
	ID       string
	Username string
	IsActive bool
}
