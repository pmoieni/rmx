{
  pkgs,
  lib,
  # config,
  # inputs,
  ...
}: {
  env.GREET = "RMX devenv";
  packages = [pkgs.git];

  scripts.hello.exec = "echo $GREET";

  # startup
  enterShell = ''
    hello
    pg_ctl status
  '';

  # tests
  # enterTest = ''
  # echo "Running tests"
  # git --version | grep "2.42.0"
  # '';

  # services
  services.postgres = {
    enable = true;
    initialScript = ''
      CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
      CREATE EXTENSION IF NOT EXISTS "citext";
    '';
    initialDatabases = [
      {
        name = "rmx-test";
        schema = ./internal/db/migrations/20240727203309_schema.up.sql;
      }
    ];
    listen_addresses = "127.0.0.1";
    port = 5432;
    settings = {
      log_connections = true;
      log_statement = "all";
      logging_collector = true;
      log_disconnections = true;
      log_destination = lib.mkForce "syslog";
    };
  };

  # languages
  languages.nix.enable = true;
  languages.go.enable = true;

  # pre-commit hooks
  # pre-commit.hooks.shellcheck.enable = true;

  # processes
  # processes.ping.exec = "ping example.com";
}
