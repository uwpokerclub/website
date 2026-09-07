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
  # one_pending_transition is an intentional manual partial index; GORM cannot
  # describe it, so exclude it from schema diffs to prevent a generated DROP.
  exclude = ["officer_transitions.one_pending_transition"]
  migration {
    dir = "file://atlas/migrations"
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
