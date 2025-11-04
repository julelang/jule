{ pkgs }:

pkgs.treefmt.withConfig {
  runtimeInputs = with pkgs; [
    nixfmt-rfc-style
    julefmt
  ];

  settings = {
    on-unmatched = "info";
    tree-root-file = "flake.nix";

    formatter = {
      julefmt = {
        command = "julefmt";
        options = [ "-w" ];
        includes = [ "*.jule" ];
      };

      nixfmt = {
        command = "nixfmt";
        includes = [ "*.nix" ];
      };
    };
  };
}
