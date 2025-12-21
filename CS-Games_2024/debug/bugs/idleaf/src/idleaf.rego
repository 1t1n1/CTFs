package idleaf

default allow := false

result := {
    "allow": allow,
    "errors": errors,
}

# Allow action if the user's role for the key supports it
allow {
    role := input.keys[input.key][input.user]
    action_permits_role(input.action, role)
}

# Collect error messages if access is denied
errors := {msg |
    not allow
    msg := sprintf("access_denied: `%s` is not allowed to perform action `%s` on key `%s`", [input.user, input.action, input.key])
}

# Helper rule to check if action is permitted by the user's role
action_permits_role(action, role) {
    role == "admin"
} else {
    action == "read"
    role == "read"
} else {
    action == "write"
    role == "write"
}

# Consider users in teams and teams within teams for access control
# This part might need adjustments based on how teams are defined and used in your input
related_entities[entity] {
    entity := input.user
} else {
    entity := input.teams[_]
}

object_roles := {entity: role |
    entity := related_entities[_]
    role := input.keys[input.key][entity]
}
