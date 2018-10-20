package secretservice

// Schema is the GraphQL schema.
const Schema = `
schema {
  query: Query
  mutation: Mutation
}

type Query {
  # release returns a single release from a particular Scope.
  release(scopeId: ID!, releaseId: ID!): Release!

  # workspace returns the current workspace for a particular Scope.
  scope(scopeId: ID!): Scope!
}

type Mutation {
  # createScope creates a new configuration scope with a given name, using the
  # provided KMS key for encryption.
  createScope(name: String!, kmsKeyId: String!): Scope!

  # addVariable adds or changes a Variable in the current workspace.
  addVariable(scopeId: ID!, variable: VariableInput!): Variable!

  # removeVariable removes a Variable from the current workspace.
  removeVariable(scopeId: ID!, id: ID!): Variable!

  # createRelease takes a snapshot of the current workspace to create a Release.
  createRelease(scopeId: ID!): Release!

  # showRelease returns a Release given its ID.
  showRelease(scopeId: ID!, releaseId: ID!): Release!

  # archiveRelease archives a Release. Archived releases should no longer be
  # available for anything other than historical purposes. This is an
  # irrevertible operation, though you can use "reset" to put the content
  # of an archived Release back in the workspace.
  archiveRelease(scopeId: ID!, releaseId: ID!): Release!

  # reset replaces the content of the current workspace with the content of
  # the Release.
  reset(scopeId: ID!, releaseId: ID!): Scope!
}

# Release is the snapshot of the configuration associated with a given Scope.
type Release {
  id: ID!
  scope: Scope!
  live: Boolean!
  variables: [Variable!]!
}

# Scope is a particular configuration scope. Configuration is available on
# per-scope basis.
type Scope {
  id: ID!
  kmsKeyId: String!
  variables: [Variable!]!
}

# Variable is a single element of the configuration.
type Variable {
  id: ID!
  value: String
  writeOnly: Boolean!
}

input VariableInput {
  name: String!
  value: String!
  writeOnly: Boolean!
}
`
