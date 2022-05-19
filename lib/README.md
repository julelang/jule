# lib

Standard library directory. <br>
The directories in this directory, accept as package. <br>
The files in this directory, accept as local package files. <br>

## Important

### Packages

Name of package directories is must conform to naming conventions of the language.

### Local Package Files

Local package files is actually direct imported defines. <br>
This files is can't imports in source code because before compilation, compiler imports these package automatically. <br>
So, not necessary. Actually, is not possible because ``use`` statement sees subpackages. <br>
For this reason, this defines, can use in everywhere and not necessary any import operation for use.
