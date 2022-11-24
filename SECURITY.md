# Jule Security Policy

The security policy of Jule.
This document discusses how security issues are handled and how they can be reported.

## Supported Versions

Jule always supports the latest version.
When there are security issues, patches for these vulnerabilities are released in the form of minor versions of the main version.

## Reporting a Vulnerability

Before reporting an issue, make sure you are using the most current version of JuleC.
If you're using an older version and newer versions don't have issues to report, it's unlikely that anything can be done about it.
Make sure the issue you are reporting is not publicly listed, if you see it publicly on the [Jule Issue Tracker](https://github.com/julelang/jule/issues) then this issue is known.

You can see all past security code names and other information in the project [security](https://github.com/orgs/julelang/projects/4).
To make it easier, you can follow all the problems here.
A public issue is shown as a direct issue, but if there is a non-public issue you will probably only see the codename.
Non-public issues become public after they are fixed.

### Classification

+ **Risky**: They are security issues that pose a risk but that can be followed publicly and that do not have a critical impact and minor or do not create significant vulnerabilities.

+ **Critical**: Critical issues are critical, high priority security issues that can disrupt the Jule ecosystem and leave all or most software in that ecosystem vulnerable.
These issues should not be publicly tracked.

### Creating Security Report

If you report using [Jule Issue Tracker](https://github.com/julelang/jule/issues), you should use the security report form.
Please don't forget to share any information you have regarding the security issue.
Reporting all the details will be helpful in solving the problem and speeding up the process.

Some security issues should not be shared publicly.
These vulnerabilities are too critical to be processed publicly, so report them to us privately to prevent them from being used for malicious purposes.
Please send an email to security@jule.dev explaining everything you know about vulnerability.

## Patch Process

First of all, when a security issue is received, the validity of the issue is checked.
If the vulnerability is valid, a codename is assigned.
The code-named issue is listed in the project [security](https://github.com/orgs/julelang/projects/4).
Then the related problems are tried to be solved. When the problem is resolved, they are published with the closest version.

## Receiving Security Updates

Security updates write their code in the release notes. If the version is released, it means that all resolved related bugs have become public.
You can reach the issues in detail from the relevant links.

## Contribute to Policy

If you have ideas to improve this security policy, please let us know on the [Jule Issue Tracker](https://github.com/julelang/jule/issues).
