package member

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleMember     Role = "member"
)

// rank 는 역할의 높낮이. 숫자가 클수록 권한이 크다.
var rank = map[Role]int{
	RoleMember:     1,
	RoleAdmin:      2,
	RoleSuperAdmin: 3,
}

// AtLeast 는 r 이 min 이상의 권한인지 본다. 예: RoleSuperAdmin.AtLeast(RoleAdmin) == true
func (r Role) AtLeast(min Role) bool {
	return rank[r] >= rank[min]
}
