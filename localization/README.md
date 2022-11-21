# Localization Configurations

Localization configurations are for the messages and logs of the JuleC compiler. <br>
They are functional in many ways, such as easily preparing translations for languages and creating alternative localization configurations.

Localization configurations are available in files. <br>
The names of the files are also the names of the localization configurations.

The files contain the content of languages in key-value format. <br>
For the display of the contents, the language elements addressed by the relevant fields must be in the INI file with a specific identifier. <br>
JuleC only tries to make sense of INI files with those identifiers.

Localization configurations  do not affect all JuleC's messages. <br>
Some messages or parts are rendered in English until the language settings are valid. <br>
However, there are content that the language settings do not affect at all.

Contents are displayed as key-value pairs. <br>
Since each key acts as a content's identity, they should never be changed. <br>
They are constant.

The keys are followed by a space and the value follows. <br>
The first space key and value distinction is accepted. <br>
The following content of the line is then rendered as a value. <br>
The leading and trailing spaces of the value are trimmed.

Each new line means a new key-value pairs. <br>
You can make empty line but if line has a content, must be key-value. <br>
If you do not specify a key-value pair, its default value will be accepted. <br>
If you try to assign a value to a key that does not exist, you will get an error. <br>
If you specify a key-value pair more than once, your most recent impression is considered.

## Arguments
Contents are processed with standard format implementations. <br>
Arguments are also processed according to the format of this library.
<br><br>

```
example_error    %s is invalid by %f ratio
```
``%s`` in the example above indicates that the first argument for that message is a string, and ``%f`` indicates that the second argument is float. <br>
If the message has no arguments, it is possible that you are not getting the desired result.
