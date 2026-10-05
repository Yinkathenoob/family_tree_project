# Family Tree Biodata Management System — Design Document

## Overview

A backend system for recording and exploring family trees. Each family tree
is a collection of people connected by parent-child relationships, from which
wider relationships (siblings, grandparents, descendants) are derived. The
system is built in Go with a PostgreSQL database and is designed to grow
feature-by-feature from a small in-memory prototype into a full HTTP API.

## Goals

- Start small (in-memory) and add a database, then an HTTP API, then a clean
  layered architecture — one new concept at a time.
- Model real family relationships, including the business rules of Nigerian
  patrilineal family structure.
- Stay modular and testable so new features can be added without rewrites.

## Functional Requirements

### Core (people)
- Add a person with their biodata.
- View a single person's details.
- List all people (optionally within one family).
- Edit a person's details.
- Delete a person (handling the case where they are referenced as a parent).

### Families
- Create a family (a named, distinct family tree).
- List families.
- A person belongs to exactly one family.

### Relationships
- Record a person's parents (mother and/or father).
- Find a person's children.
- Find a person's siblings.
- Find a person's parents and grandparents.
- Find a person's descendants and ancestors.

### Deferred / future
- Search people by name.
- Record marriages/spouses (introduces birth-family vs marital-family).
- Richer adoption modelling (adopted by one parent, biological to another).
- Export or visualise the tree.

## Non-Functional Requirements

- **Scalability (of features):** modular architecture; new capabilities added
  without rewriting existing ones.
- **Maintainability:** clear separation of concerns via layered architecture.
- **Data integrity:** parent references must point to real people; a person's
  birth date cannot be after a parent's death in impossible ways; referential
  rules enforced by the database (foreign keys, constraints).
- **Persistence:** data survives restarts (PostgreSQL).
- **Testability:** each layer testable in isolation; business logic separated
  from I/O.

## Data Model

### Entities

**Family**

| Attribute | Type | Notes |
|---|---|---|
| ID | integer | unique, auto-assigned |
| Name | text | e.g. "The Adeyemi Family" |
| Notes | text (nullable) | origin, context |

**Person**

| Attribute | Type | Notes |
|---|---|---|
| ID | integer | unique, auto-assigned |
| FamilyID | integer (FK → Family, nullable) | current/primary family; for a wife, the husband's family. NULL only for an unplaced root |
| BirthFamilyID | integer (FK → Family, nullable) | family the person was born into; differs from FamilyID for someone who married in |
| FirstName | text | |
| LastName | text | |
| Gender | text | |
| BirthDate | date (nullable) | exact date; optional (may be unknown) |
| DeathDate | date (nullable) | empty = living |
| Address | text (nullable) | single address for now |
| MotherID | integer (FK → Person, nullable) | empty for roots |
| FatherID | integer (FK → Person, nullable) | empty for roots |
| Adopted | boolean | simple flag for now |
| Notes | text (nullable) | optional |

### Relationships
- **Family → Person:** one-to-many. A family has many people; each person
  belongs to one family.
- **Person → Person (parent):** self-referential. A person references their
  mother and father, who are also people. All wider relationships (siblings,
  grandparents, descendants) are *derived* from these links, not stored.

### Design principle
Store the minimal source of truth (parent links, family membership) and derive
everything else (siblings, ancestors, descendants) by querying. Avoid storing
relationships that can be computed, to prevent sync errors.

## Business Rules

1. **Patrilineal family inheritance (enforced):** if a father is recorded, the
   child's FamilyID is always set to the father's FamilyID. This is forced, not
   a suggestion — it cannot be overridden while a father is present.
2. **Fatherless fallback:** if no father is recorded, FamilyID falls back to the
   mother's FamilyID. If neither parent is recorded (a tree root), FamilyID may
   be left unset (NULL) or set explicitly. In practice a tree's founder is given
   an explicit FamilyID when the family is created; NULL is mainly for a person
   added before their family is decided.
3. **Wife joins husband's family:** a woman who marries into a family becomes
   part of her husband's family. Her FamilyID reflects her marital (husband's)
   family. Her original birth family is recorded separately for reference (see
   below).

### Birth family
- A `BirthFamilyID` (nullable) records the family a person was born into, kept
  distinct from `FamilyID` (their current/primary family). For most people born
  into the tree these are the same; they differ for a wife who married in, whose
  `FamilyID` is the husband's family and whose `BirthFamilyID` is her own.

### Deferred rules
- **Full marriage/spouse modelling:** recording marriages as their own
  relationship (spouse links, marriage dates, divorce) is deferred. For now only
  the family-membership effect of marriage (wife joins husband's family, birth
  family retained) is captured via `FamilyID` and `BirthFamilyID`.

## Architecture

The system is built in layers, each with a single responsibility, connected
through interfaces so each can change or be tested independently.

- **Handler / API layer:** receives HTTP requests, returns responses. Knows
  HTTP; knows nothing about the database.
- **Service / business-logic layer:** the rules and operations (add a person,
  apply the patrilineal rule, find siblings). Knows nothing about HTTP or the
  database — pure logic.
- **Repository / data layer:** talks to PostgreSQL, runs queries. Knows the
  database; knows nothing about HTTP.

The service depends on a *repository interface*, not a concrete database, so a
fake repository can be substituted in tests (dependency injection).

## Build Stages

1. **In-memory prototype:** Person/Family structs, held in slices, with
   functions to add people and compute relationships. No database, no HTTP.
2. **PostgreSQL:** real tables, connect from Go, move storage into the database.
3. **HTTP API:** expose operations as REST endpoints.
4. **Layered refactor:** separate handler / service / repository with interfaces.
5. **Hardening:** validation, authentication, pagination, migrations, tests.
