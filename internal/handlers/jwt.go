package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// TokenClaims representa os claims do token JWT
type TokenClaims struct {
	Sub           string          `json:"sub"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	Role          string          `json:"role"`
	MsRole        json.RawMessage `json:"http://schemas.microsoft.com/ws/2008/06/identity/claims/role"`
	MsRoles       []string        `json:"-"` // Campo auxiliar para armazenar MsRole quando é um array
	Roles         []string        `json:"roles"`
	Groups        []string        `json:"group"`
	PreferredName string          `json:"preferred_username"`
}

// ParseJWTClaims extrai os claims de um token JWT
func ParseJWTClaims(tokenString string) (*TokenClaims, error) {
	// Divide o token em 3 partes (header, payload, signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token inválido: formato incorreto")
	}

	// Decodifica a parte do payload (claims)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar payload: %w", err)
	}

	// Parse do JSON para struct
	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("erro ao analisar claims: %w", err)
	}

	// Processa MsRole que pode ser string ou array
	if len(claims.MsRole) > 0 {
		var msRoleString string
		if err := json.Unmarshal(claims.MsRole, &msRoleString); err == nil {
			// É uma string
			claims.MsRoles = []string{msRoleString}
			log.Printf("Processado MsRole como string: %s", msRoleString)
		} else {
			// Tenta como array
			var msRoleArray []string
			if err := json.Unmarshal(claims.MsRole, &msRoleArray); err == nil {
				claims.MsRoles = msRoleArray
				log.Printf("Processado MsRole como array: %v", msRoleArray)
			} else {
				log.Printf("Erro ao processar MsRole: %v, valor raw: %s", err, string(claims.MsRole))
			}
		}
	}

	return &claims, nil
}

// HasValidRole verifica se o usuário tem pelo menos uma das roles permitidas
func HasValidRole(claims *TokenClaims) bool {
	// Lista de roles permitidas
	validRoles := []string{"Administrator", "Consultant", "Admin", "admin", "IdentityConsultant"}

	log.Printf("Verificando roles para usuário %s: Role padrão: [%s], MsRoles: %v, Roles: %v, Groups: %v",
		claims.Name, claims.Role, claims.MsRoles, claims.Roles, claims.Groups)

	// Verificar se o usuário tem pelo menos uma role (independente de qual for)
	hasAnyRole := false
	if claims.Role != "" {
		hasAnyRole = true
	}
	if len(claims.MsRoles) > 0 {
		hasAnyRole = true
	}
	if len(claims.Roles) > 0 {
		hasAnyRole = true
	}
	if len(claims.Groups) > 0 {
		hasAnyRole = true
	}

	// Se o usuário não tem nenhuma role, negar acesso
	if !hasAnyRole {
		log.Printf("Usuário %s não tem nenhuma role atribuída", claims.Name)
		return false
	}

	// Verifica se o campo role existe e é válido
	if claims.Role != "" {
		for _, validRole := range validRoles {
			if strings.EqualFold(claims.Role, validRole) {
				log.Printf("Role válida encontrada em claims.Role: %s", claims.Role)
				return true
			}
		}

		// Verificação especial para "IdentityConsultant" no campo Role
		if strings.Contains(strings.ToLower(claims.Role), "consultant") {
			log.Printf("String 'consultant' encontrada em claims.Role: %s", claims.Role)
			return true
		}
	}

	// Verifica nas MsRoles processadas
	for _, msRole := range claims.MsRoles {
		for _, validRole := range validRoles {
			if strings.EqualFold(msRole, validRole) {
				log.Printf("Role válida encontrada em claims.MsRoles: %s", msRole)
				return true
			}
		}

		// Verificação especial para "IdentityConsultant" que equivale a "Consultant"
		if strings.Contains(strings.ToLower(msRole), "consultant") ||
			strings.Contains(strings.ToLower(msRole), "identity") {
			log.Printf("String 'consultant' ou 'identity' encontrada em claims.MsRoles: %s", msRole)
			return true
		}
	}

	// Verifica no array de roles
	for _, role := range claims.Roles {
		for _, validRole := range validRoles {
			if strings.EqualFold(role, validRole) {
				log.Printf("Role válida encontrada em claims.Roles: %s", role)
				return true
			}
		}

		// Verificação especial para "IdentityConsultant" no array de roles
		if strings.Contains(strings.ToLower(role), "consultant") ||
			strings.Contains(strings.ToLower(role), "identity") {
			log.Printf("String 'consultant' ou 'identity' encontrada em claims.Roles: %s", role)
			return true
		}
	}

	// Verifica no array de grupos (algumas vezes as roles vêm como grupos)
	for _, group := range claims.Groups {
		if strings.Contains(strings.ToLower(group), "consultant") ||
			strings.Contains(strings.ToLower(group), "admin") ||
			strings.Contains(strings.ToLower(group), "identity") {
			log.Printf("String 'consultant', 'admin' ou 'identity' encontrada em claims.Groups: %s", group)
			return true
		}
	}

	log.Printf("Nenhuma role válida encontrada para o usuário %s", claims.Name)
	return false
}

// DumpClaimsInfo registra todas as informações de claims para depuração
func DumpClaimsInfo(claims *TokenClaims) {
	log.Printf("=== DUMP DE CLAIMS PARA USUÁRIO: %s ===", claims.Name)
	log.Printf("Sub: %s", claims.Sub)
	log.Printf("Name: %s", claims.Name)
	log.Printf("Email: %s", claims.Email)
	log.Printf("Role: %s", claims.Role)
	log.Printf("MsRoles raw: %s", string(claims.MsRole))
	log.Printf("MsRoles processado: %v", claims.MsRoles)
	log.Printf("Roles: %v", claims.Roles)
	log.Printf("Groups: %v", claims.Groups)
	log.Printf("PreferredName: %s", claims.PreferredName)

	// Verificar a existência de "IdentityConsultant" em todos os campos
	if strings.EqualFold(claims.Role, "IdentityConsultant") {
		log.Printf("Role 'IdentityConsultant' encontrada exata em claims.Role")
	}

	for _, role := range claims.MsRoles {
		if strings.EqualFold(role, "IdentityConsultant") {
			log.Printf("Role 'IdentityConsultant' encontrada exata em claims.MsRoles")
		}
	}

	for _, role := range claims.Roles {
		if strings.EqualFold(role, "IdentityConsultant") {
			log.Printf("Role 'IdentityConsultant' encontrada exata em claims.Roles")
		}
	}

	for _, group := range claims.Groups {
		if strings.EqualFold(group, "IdentityConsultant") {
			log.Printf("Role 'IdentityConsultant' encontrada exata em claims.Groups")
		}
	}

	log.Printf("=== FIM DO DUMP DE CLAIMS ===")
}
