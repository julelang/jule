# Language Packs

Language packs are for the messages and logs of the XXC compiler. <br>
They are functional in many ways, such as easily preparing translations for languages and creating alternative language packs.

Language packs are available in directories. <br>
The names of the directories are also the names of the language packs.

The directories contain the content of languages in JSON format. <br>
For the display of the contents, the language elements addressed by the relevant fields must be in the JSON file with a specific name. <br>
XXC only tries to make sense of JSON files with those names. <br>
It's okay if language packs contain directories or different files.

Language packs do not affect all XXC's messages. <br>
Some messages or parts are rendered in English until the language settings are valid. <br>
However, there are content that the language settings do not affect at all.

Contents are displayed in JSON as key-value pairs. <br>
Since each key acts as a content's identity, they should never be changed. <br>
They are constant.

## File Names and Contents

The following table lists the specially named files that language packs must contain and the purposes of their contents. <br>
If the language pack does not have any of these files, the default language elements are used for missing places.

<table>
  <tr>
    <td><strong>File Name</strong></td>
    <td><strong>Content</strong></td>
  </tr>
  <tr>
    <td>errs.json</td>
    <td>Error messages</td>
  </tr>
  <tr>
    <td>warns.json</td>
    <td>Warning messages</td>
  </tr>
</table>

## Arguments
Contents are processed with standard format implementations. <br>
Arguments are also processed according to the format of this library.
<br><br>

```json
"example_error": "%s is invalid by %f ratio"
```
``%s`` in the example above indicates that the first argument for that message is a string, and ``%f`` indicates that the second argument is float. <br>
If the message has no arguments, it is possible that you are not getting the desired result.
