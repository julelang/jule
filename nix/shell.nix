{ pkgs }:

let
  mainPkg = pkgs.callPackage ./default.nix { };
  packages =
    with pkgs;
    [
      clang
      julec
      julefmt
      typos
    ]
    ++ mainPkg.nativeBuildInputs;
in
pkgs.mkShell {
  inherit packages;

  shellHook = ''
    echo -ne "------------------------------------\n "

    echo -n "${toString (map (pkg: "• ${pkg.name}\n") packages)}"

    echo "------------------------------------"
  '';
}
