data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/models",
    "--dialect", "postgres", // | postgres | sqlite | sqlserver
  ]
}
env "gorm" {
  src = data.external_schema.gorm.url
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/17/dev?search_path=public"
  migration {
    dir = "file://atlas/migrations"
    # These objects are defined in released SQL but Atlas's GORM schema cannot
    # describe them faithfully. The relationship copies logins.username's text
    # type onto account_activations.username, although the released child column
    # is unbounded varchar.
    # GORM also cannot describe users_quest_id_unique's partial expression index;
    # it is created by manual SQL and must be excluded to prevent generated drops.
    exclude = ["*.one_pending_transition", "account_activations.username", "*.users_quest_id_unique"]
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
  diff {
    concurrent_index {
      create = true
      drop   = true
    }
  }
}
