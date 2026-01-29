package cmd_user

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Rafael24595/go-api-core/src/application/command/apps"
	"github.com/Rafael24595/go-api-core/src/application/session"
	"github.com/Rafael24595/go-api-core/src/commons/utils"
	domain_session "github.com/Rafael24595/go-api-core/src/domain/session"
	"github.com/Rafael24595/go-collections/collection"
)

const Command apps.SnapshotFlag = "user"

const (
	FLAG_HELP         = "-h"
	FLAG_VERBOSE      = "-v"
	FLAG_LOCK         = "-lk"
	FLAG_UNLOCK       = "-uk"
	FLAG_ROLE_LIST    = "-r"
	FLAG_USER_LIST    = "-u"
	FLAG_USER_DETAILS = "-d"
	FLAG_USER_PUSH    = "-p"
	FLAG_ADD_ROLE     = "-ar"
	FLAG_USER_REMOVE  = "-rm"
)

const (
	TABLE_USERNAME  = "Username"
	TABLE_LOCK      = "Lock"
	TABLE_DATE      = "Date"
	TABLE_PUBLISHER = "Publisher"
	TABLE_ROLES     = "Roles"
	TABLE_ACCESS    = "Access"
)

var App = apps.CommandApplication{
	CommandReference: apps.CommandReference{
		Flag:        Command,
		Name:        "User",
		Description: "Manages system users",
		Example:     refHelp.Example,
	},
	Exec: exec,
	Help: help,
}

var refs = []apps.CommandReference{
	refHelp,
	refRoleList,
	refUserList,
	refUserDetails,
	refUserPush,
	refUserRemove,
	refUserLock,
	refUserUnlock,
}

var refHelp = apps.CommandReference{
	Flag:        FLAG_HELP,
	Name:        "Help",
	Description: "Shows this help message.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_HELP),
}

var refRoleList = apps.CommandReference{
	Flag:        FLAG_ROLE_LIST,
	Name:        "Role list",
	Description: "Displays the list of public roles.",
	Example:     fmt.Sprintf(`%s %s`, Command, FLAG_ROLE_LIST),
}

var refUserList = apps.CommandReference{
	Flag:        FLAG_USER_LIST,
	Name:        "User list",
	Description: "Displays the list of users, use the verbose flag to expand the details.",
	Example:     fmt.Sprintf(`%s %s %s`, Command, FLAG_VERBOSE, FLAG_USER_LIST),
}

var refUserDetails = apps.CommandReference{
	Flag:        FLAG_USER_DETAILS,
	Name:        "Details",
	Description: "Displays the data for a given user.",
	Example:     fmt.Sprintf(`%s %s ${username}`, Command, FLAG_USER_DETAILS),
}

var refUserPush = apps.CommandReference{
	Flag:        FLAG_USER_PUSH,
	Name:        "Push",
	Description: "Insert a new user.",
	Example:     fmt.Sprintf(`%s %s ${user}#${pass} %s ${role}+${role}`, Command, FLAG_USER_PUSH, FLAG_ADD_ROLE),
}

var refUserRemove = apps.CommandReference{
	Flag:        FLAG_USER_REMOVE,
	Name:        "Remove",
	Description: "Remove the given user.",
	Example:     fmt.Sprintf(`%s %s ${user}`, Command, FLAG_USER_REMOVE),
}

var refUserLock = apps.CommandReference{
	Flag:        FLAG_LOCK,
	Name:        "Lock",
	Description: "Lock the given user.",
	Example:     fmt.Sprintf(`%s %s ${user}+${user}`, Command, FLAG_LOCK),
}

var refUserUnlock = apps.CommandReference{
	Flag:        FLAG_UNLOCK,
	Name:        "Lock",
	Description: "Lock the given user.",
	Example:     fmt.Sprintf(`%s %s ${user}+${user}`, Command, FLAG_UNLOCK),
}

type pushPayload struct {
	name  string
	pass  string
	roles []domain_session.Role
}

func exec(request *apps.CmdExecRequest) *apps.CmdExecResult {
	user := request.User
	cmd := request.Command

	if cmd.Size() == 0 {
		return help()
	}

	sess := session.InstanceManagerSession()
	publ, err := sess.FindProvider(user)
	if err != nil {
		return apps.ErrorResult(err)
	}

	verbose := false

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_HELP:
			return help()
		case FLAG_VERBOSE:
			verbose = true
		case FLAG_ROLE_LIST:
			return roleslist()
		case FLAG_USER_LIST:
			tuple, err := apps.ResolveKeyValueCursor(cmd, "=", false)
			if err != nil {
				return apps.ErrorResult(err)
			}
			return execUserList(cmd, tuple, verbose)
		case FLAG_USER_DETAILS:
			return userDetails(cmd)
		case FLAG_USER_PUSH:
			re := regexp.MustCompile(`#([^\s]+)`)
			hidden := re.ReplaceAllString(request.Input, "#****")

			tuple, err := apps.ResolveKeyValueCursor(cmd, "#", true)
			if err != nil {
				return apps.ErrorResult(err).SetInput(hidden)
			}

			return execUserPush(publ, cmd, tuple.Flag, tuple.Data).SetInput(hidden)
		case FLAG_USER_REMOVE:
			users, err := apps.ResolveChainCursor(cmd, "+")
			if err != nil {
				return apps.ErrorResult(err)
			}
			return execUserRemove(publ, cmd, users)
		case FLAG_LOCK:
			result, err := apps.ResolveChainCursor(cmd, "+")
			if err != nil {
				return apps.ErrorResult(err)
			}
			return execUserStatus(publ, cmd, result, make([]string, 0))
		case FLAG_UNLOCK:
			result, err := apps.ResolveChainCursor(cmd, "+")
			if err != nil {
				return apps.ErrorResult(err)
			}
			return execUserStatus(publ, cmd, make([]string, 0), result)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return apps.EmptyResult()
}

func help() *apps.CmdExecResult {
	title := fmt.Sprintf("Available %s actions:\n", Command)
	return apps.RunHelp(title, refs)
}

func roleslist() *apps.CmdExecResult {
	sess := session.InstanceManagerSession()

	buffer := make([]string, 0)

	buffer = append(buffer, "Public user roles:")

	for _, v := range sess.GetPublicRoles() {
		buffer = append(buffer, fmt.Sprintf(" - %s", v))
	}

	return apps.NewResult(strings.Join(buffer, "\n"))
}

func execUserList(cmd *collection.Vector[string], tuple *utils.CmdTuple, verbose bool) *apps.CmdExecResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_VERBOSE:
			verbose = true
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	result := userlist(tuple, verbose)

	return apps.NewResult(result)
}

func userlist(tuple *utils.CmdTuple, verbose bool) string {
	sess := session.InstanceManagerSession()
	users := collection.VectorFromList(sess.FindAll())

	if tuple != nil {
		switch tuple.Flag {
		case "name":
			users.FilterSelf(func(u domain_session.SessionLite) bool {
				return strings.Contains(u.Username, tuple.Data)
			})
		}
	}

	users.Sort(func(i, j domain_session.SessionLite) bool {
		return i.Timestamp < j.Timestamp
	})

	if verbose {
		return formatSessionsVerbose(users.Collect()...)
	}

	return formatSessions(users)
}

func userDetails(cmd *collection.Vector[string]) *apps.CmdExecResult {
	value, err := apps.ResolveValueCursor(cmd)
	if err != nil {
		return apps.ErrorResult(err)
	}

	sess := session.InstanceManagerSession()

	s, ok := sess.Find(value)
	if !ok {
		return apps.NewResultf("user '%s' not found", value)
	}

	result := formatSessionsVerbose(*domain_session.ToLite(*s))

	return apps.NewResult(result)
}

func execUserPush(publ *domain_session.Session, cmd *collection.Vector[string], name, pass string) *apps.CmdExecResult {
	pushData := pushPayload{
		name:  name,
		pass:  pass,
		roles: make([]domain_session.Role, 0),
	}

	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_ADD_ROLE:
			roles, err := resolveRoles(cmd)
			if err != nil {
				return apps.ErrorResult(err)
			}
			pushData.roles = append(pushData.roles, roles...)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return userPush(publ, pushData)
}

func userPush(publ *domain_session.Session, payload pushPayload) *apps.CmdExecResult {
	sess := session.InstanceManagerSession()

	if _, ok := sess.Find(payload.name); ok {
		return apps.NewResultf("user '%s' already exists", payload.name)
	}

	user, err := sess.Insert(publ, payload.name, payload.pass, payload.roles)
	if err != nil {
		return apps.ErrorResult(err)
	}

	return apps.NewResultf("user '%s' created successfully", user.Username)
}

func execUserRemove(publ *domain_session.Session, cmd *collection.Vector[string], users []string) *apps.CmdExecResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_USER_REMOVE:
			result, err := apps.ResolveChainCursor(cmd, "+")
			if err != nil {
				return apps.ErrorResult(err)
			}
			users = append(users, result...)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	return userRemove(publ, users)
}

func userRemove(publ *domain_session.Session, names []string) *apps.CmdExecResult {
	sess := session.InstanceManagerSession()

	users := make([]domain_session.Session, 0)
	for _, v := range names {
		user, ok := sess.Find(v)
		if !ok {
			return apps.NewResultf("user '%s' doesnt exists", v)
		}

		if user.HasRole(domain_session.ROLE_PROTECTED) {
			return apps.NewResultf("user '%s' cannot be removed", v)
		}

		users = append(users, *user)
	}

	removed := make([]string, 0)
	for _, v := range users {
		user, err := sess.Delete(publ, &v)
		if err != nil {
			return apps.ErrorResult(err)
		}
		removed = append(removed, user.Username)
	}

	return apps.NewResultf("%d users removed: %s", len(removed), strings.Join(removed, " "))
}

func execUserStatus(publ *domain_session.Session, cmd *collection.Vector[string], usersLock, usersUnlock []string) *apps.CmdExecResult {
	for cmd.Size() > 0 {
		flag, ok := cmd.Shift()
		if !ok {
			break
		}

		switch flag {
		case FLAG_LOCK:
			result, err := apps.ResolveChainCursor(cmd, "+")
			if err != nil {
				return apps.ErrorResult(err)
			}
			usersLock = append(usersLock, result...)
		case FLAG_UNLOCK:
			result, err := apps.ResolveChainCursor(cmd, "+")
			if err != nil {
				return apps.ErrorResult(err)
			}
			usersUnlock = append(usersUnlock, result...)
		default:
			return apps.NewResultf("Unrecognized command flag: %s", flag)
		}
	}

	buffer := make([]string, 0)

	if message := userStatus(publ, true, usersLock); message != "" {
		buffer = append(buffer, message)
	}

	if message := userStatus(publ, false, usersUnlock); message != "" {
		buffer = append(buffer, message)
	}

	return apps.NewResult(strings.Join(buffer, "\n"))
}

func userStatus(publ *domain_session.Session, status bool, names []string) string {
	sess := session.InstanceManagerSession()

	users := make([]domain_session.Session, 0)
	for _, v := range names {
		user, ok := sess.Find(v)
		if !ok {
			return fmt.Sprintf("user '%s' doesnt exists", v)
		}

		if user.HasRole(domain_session.ROLE_PROTECTED) {
			return fmt.Sprintf("user '%s' cannot be removed", v)
		}

		users = append(users, *user)
	}

	updated := make([]string, 0)

	for _, v := range users {
		var user *domain_session.Session
		var err error

		if status {
			user, err = sess.Unlock(publ, &v)
		} else {
			user, err = sess.Lock(publ, &v)
		}

		if err != nil {
			return err.Error()
		}

		updated = append(updated, user.Username)
	}

	if len(updated) == 0 {
		return ""
	}

	strStatus := "locked"
	if status {
		strStatus = "unlocked"
	}

	return fmt.Sprintf("%d users %s: %s", len(updated), strStatus, strings.Join(updated, " "))
}

func formatSessionsVerbose(users ...domain_session.SessionLite) string {
	table := utils.NewTable()

	table.Headers(TABLE_USERNAME, TABLE_LOCK, TABLE_DATE, TABLE_PUBLISHER, TABLE_ROLES, TABLE_ACCESS)

	for i, v := range users {
		table.Field(TABLE_USERNAME, i, v.Username)
		table.Field(TABLE_LOCK, i, v.Lock)
		table.Field(TABLE_DATE, i, utils.FormatMilliseconds(v.Timestamp))
		table.Field(TABLE_PUBLISHER, i, v.Publisher)
		table.Field(TABLE_ROLES, i, v.Roles)
		table.Field(TABLE_ACCESS, i, v.Count)
	}

	return table.ToString()
}

func formatSessions(users *collection.Vector[domain_session.SessionLite]) string {
	_, maxlen, _ := users.Max(func(i domain_session.SessionLite) int {
		return len(i.Username)
	})

	return collection.VectorMap(users,
		func(s domain_session.SessionLite) string {
			space := strings.Repeat(" ", maxlen-len(s.Username))
			return fmt.Sprintf(" %s%s   %s", s.Username, space, utils.FormatMilliseconds(s.Timestamp))
		}).Join("\n")
}

func resolveRoles(cmd *collection.Vector[string]) ([]domain_session.Role, error) {
	value, err := apps.ResolveChainCursor(cmd, "+")
	if err != nil {
		return make([]domain_session.Role, 0), err
	}

	roles := make([]domain_session.Role, 0)

	for _, v := range value {
		role, ok := domain_session.RoleFromString(v)
		if !ok {
			return roles, fmt.Errorf("Role '%s' does not exist", v)
		}

		roles = append(roles, role)
	}

	return roles, nil
}
