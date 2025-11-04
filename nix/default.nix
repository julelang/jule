{
  julec,
}:

julec.overrideAttrs (old: {
  src = ../.;
  passthru = { };
})
