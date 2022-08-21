# Language Packs

Language packs are for the messages and logs of the JuleC compiler. <br>
They are functional in many ways, such as easily preparing translations for languages and creating alternative language packs.

Language packs are available in files. <br>
The names of the files are also the names of the language packs.

The files contain the content of languages in JSON format. <br>
For the display of the contents, the language elements addressed by the relevant fields must be in the JSON file with a specific identifier. <br>
JuleC only tries to make sense of JSON files with those identifiers.

Language packs do not affect all JuleC's messages. <br>
Some messages or parts are rendered in English until the language settings are valid. <br>
However, there are content that the language settings do not affect at all.

Contents are displayed in JSON as key-value pairs. <br>
Since each key acts as a content's identity, they should never be changed. <br>
They are constant.

## Arguments
Contents are processed with standard format implementations. <br>
Arguments are also processed according to the format of this library.
<br><br>

```json
"example_error": "%s is invalid by %f ratio"
```
``%s`` in the example above indicates that the first argument for that message is a string, and ``%f`` indicates that the second argument is float. <br>
If the message has no arguments, it is possible that you are not getting the desired result.
