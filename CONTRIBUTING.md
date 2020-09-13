# Contributing to secret-manager

:+1::tada: First off, thanks for taking the time to contribute! :tada::+1:

The following is a set of guidelines for contributing to secret-manager, which are hosted in the [itscontained Organization](https://github.com/itscontained) on GitHub. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

#### Table Of Contents

[Code of Conduct](#code-of-conduct)

[I don't want to read this whole thing, I just have a question!!!](#i-dont-want-to-read-this-whole-thing-i-just-have-a-question)

[How Can I Contribute?](#how-can-i-contribute)
  * [Reporting Bugs](#reporting-bugs)
  * [Suggesting Enhancements](#suggesting-enhancements)
  * [Your First Code Contribution](#your-first-code-contribution)
  * [Pull Requests](#pull-requests)

[Style Guides](#style-guides)
  * [Git Commit Messages](#git-commit-messages)
  * [Golang Styleguide](#golang-style-guide)
  * [Specs Styleguide](#specs-styleguide)
  * [Documentation Styleguide](#documentation-styleguide)

[Additional Notes](#additional-notes)
  * [Issue and Pull Request Labels](#issue-and-pull-request-labels)

## Code of Conduct

This project and contributors in it are governed by the [itscontained Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [itscontained-conduct@gmail.com](mailto:itscontained-conduct@gmail.com).

## I don't want to read this whole thing I just have a question!!!

> **Note:** [Please don't file an issue to ask a question.](http://#) You'll get faster results by using the resources below.

We are working on a detailed FAQ and where the community can chime in with helpful advice if you have questions.
* [itscontained FAQ](https://secret-manager.itscontained.io/faq)

If chat is more your speed, you can join the itscontained Discord:

* [Discord, the official itscontained server](https://discord.com/eT6crpT)
    * Even though Discord is a chat service, sometimes it takes several hours for community members to respond &mdash; please be patient!
    * Use the `#general` channel for general questions or discussion about itscontained projects
    * Use the `#secret-manager` channel for questions about secret-manager
    * Use the `#packages` channel for questions or discussion about writing or contributing to itscontained packages (both core and community)
    * Use the `#docker` channel for questions and discussion about itscontained docker containers
    * There are many other channels available, check the channel list

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for secret-manager. Following these guidelines helps maintainers, and the community, understand your report :pencil:, reproduce the behavior :computer: :computer:, and find related reports :mag_right:.

Before creating bug reports, please check [this list](#before-submitting-a-bug-report) as you might find out that you don't need to create one. When you are creating a bug report, please [include as many details as possible](#how-do-i-submit-a-good-bug-report). Fill out [the required template](https://github.com/itscontained/secret-manager/.github/blob/master/.github/ISSUE_TEMPLATE/bug_report.md), the information it asks for helps us resolve issues faster.

> **Note:** If you find a **Closed** issue that seems like it is the same thing that you're experiencing, open a new issue and include a link to the original issue in the body of your new one.

#### Before Submitting A Bug Report

* **Check the [FAQs](https://secret-manager.itscontained.io/faq)** for a list of common questions and problems.
* **Perform a [cursory search](https://github.com/search?q=is%3Aissue+user%3Aitscontained++repo%3Asecret-manager&type=issues)** to see if the problem has already been reported. If it has, **and the issue is still open**, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Bug Report?

We track bugs as [GitHub issues](https://guides.github.com/features/issues/). Create an issue and provide the following information by filling in [the template](https://github.com/itscontained/secret-manager/.github/blob/master/.github/ISSUE_TEMPLATE/bug_report.md).

Explain the problem and include additional details to help maintainers reproduce the problem:

* **Use a clear and descriptive title** for the issue to identify the problem.
* **Describe the exact steps which reproduce the problem** in as many details as possible. For example, start by explaining how you deployed secret-manager, e.g. via Helm with `<these values>`, or how you deployed secret-manager otherwise. When listing steps, **don't just say what you did, but explain how you did it**. For example, if you added an ExternalSecret, explain if you used kubectl apply, or an automation, and if so which one?
* **Provide specific examples to demonstrate the steps**. Include links to files or GitHub projects, or copy/paste-able snippets, which you use in those examples. If you're providing snippets in the issue, use [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
* **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
* **Explain which behavior you expected to see instead and why.**
* **Include screenshots and animated GIFs** which show you following the described steps and clearly demonstrate the problem. If you use the keyboard while following the steps, **record the GIF with the [Keybinding Resolver](https://github.com/atom/keybinding-resolver) shown**. You can use [this tool](https://www.cockos.com/licecap/) to record GIFs on macOS and Windows, and [this tool](https://github.com/colinkeenan/silentcast) or [this tool](https://github.com/GNOME/byzanz) on Linux.
* **If you're reporting that secret-manager crashed**, include a crash report with a stack trace from the logs. Include the crash report in the issue in a [code block](https://help.github.com/articles/markdown-basics/#multiple-lines), a [file attachment](https://help.github.com/articles/file-attachments-on-issues-and-pull-requests/), or put it in a [gist](https://gist.github.com/) and provide a link to that gist.

Provide more context by answering these questions:

* **Did the problem start happening recently** (e.g. after updating to a new version of secret-manager) or was this always a problem?
* If the problem started happening recently, **can you reproduce the problem in an older version of secret-manager?** What's the most recent version in which the problem does not happen? You can download older versions of secret-manager from [the releases page](https://github.com/itscontained/secret-manager/releases).
* **Can you reliably reproduce the issue?** If not, provide details about how often the problem happens and under which conditions it normally happens.
* If the problem relates to working with a store backend (e.g. aws or vault), **does the problem happen for all configurations or only some?**

Include details about your configuration and environment:

* **Which version of secret-manager are you using?** You can get the exact version by running `kubectl get pod  -l app.kubernetes.io/instance=secret-manager -o json | jq -r '.items[0].spec.containers[0].image'` in your kubernetes context.
* **What's the name and version of kubernetes you're using**?

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for Atom, including completely new features and minor improvements to existing functionality. Following these guidelines helps maintainers, and the community, understand your suggestion :pencil: and find related suggestions :mag_right:.

When you are creating an enhancement suggestion, please [include as many details as possible](#how-do-i-submit-a-good-enhancement-suggestion). Fill in [the template](https://github.com/itscontained/secret-manager/.github/blob/master/.github/ISSUE_TEMPLATE/feature_request.md), including the steps that you imagine you would take if the feature you're requesting existed.

#### How Do I Submit A (Good) Enhancement Suggestion?

Enhancement suggestions track as [GitHub issues](https://guides.github.com/features/issues/). Create an issue on that repository and provide the following information:

* **Use a clear and descriptive title** for the issue to identify the suggestion.
* **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
* **Provide specific examples to demonstrate the steps**. Include copy/paste-able snippets which you use in those examples, as [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
* **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
* **Include screenshots and animated GIFs** which help you demonstrate the steps or point out the part of secret-manager which the suggestion relates to. You can use [this tool](https://www.cockos.com/licecap/) to record GIFs on macOS and Windows, and [this tool](https://github.com/colinkeenan/silentcast) or [this tool](https://github.com/GNOME/byzanz) on Linux.
* **Explain why this enhancement would be useful** to most secret-manager users.
* **List some other external-secret-syncing projects where this enhancement exists.**
* **Specify which version of secret-manager you're using.** You can get the exact version by running `kubectl get pod  -l app.kubernetes.io/instance=secret-manager -o json | jq -r '.items[0].spec.containers[0].image'` in your kubernetes context.
* **Specify the name and version of kubernetes you're using.**

### Your First Code Contribution

Unsure where to begin contributing to Atom? You can start by looking through these `beginner` and `help-wanted` issues:

* [Beginner issues][beginner] - issues which should only require a few lines of code, and a test or two.
* [Help wanted issues][help-wanted] - issues which should be a bit more involved than `beginner` issues.

Both issue lists sort by total number of comments. While not perfect, number of comments is a reasonable proxy for impact a given change will have.

#### Local development

secret-manager can be developed locally. For instructions on how to do this, see the following sections in the [secret-manager docs](https://#):

### Pull Requests

The process described here has several goals:

- Maintain secret-manager's quality
- Fix problems that are important to users
- Engage the community in working toward the best possible secret-manager
- Enable a sustainable system for secret-manager's maintainers to review contributions

Please follow these steps to have your contribution considered by the maintainers:

1. Follow all instructions in [the template](.github/PULL_REQUEST_TEMPLATE.md)
2. Follow the [Style Guides](#style-guides)
3. After you submit your pull request, verify that all [status checks](https://help.github.com/articles/about-status-checks/) are passing <details><summary>What if the status checks are failing?</summary>If a status check is failing, and you believe that the failure is unrelated to your change, please leave a comment on the pull request explaining why you believe the failure is unrelated. A maintainer will re-run the status check for you. If we conclude that the failure was a false positive, then we will open an issue to track that problem with our status check suite.</details>

While the prerequisites above must be satisfied prior to having your pull request reviewed, the reviewer(s) may ask you to complete additional design work, tests, or other changes before your pull request can be ultimately accepted.

## Style Guides

### Git Commit Messages

* Use the present tense ("Feature" not "Added feature")
* Use the imperative mood ("Move to..." not "Moves to...")
* Limit the first line to 72 characters or fewer
* Reference issues and pull requests liberally after the first line
* When only changing documentation, include `[ci skip]` in the commit title
* Consider starting the commit message with an applicable emoji:
    * :art: `:art:` when improving the format/structure of the code
    * :racehorse: `:racehorse:` when improving performance
    * :non-potable_water: `:non-potable_water:` when plugging memory leaks
    * :memo: `:memo:` when writing docs
    * :penguin: `:penguin:` when fixing something on Linux
    * :apple: `:apple:` when fixing something on macOS
    * :checkered_flag: `:checkered_flag:` when fixing something on Windows
    * :bug: `:bug:` when fixing a bug
    * :fire: `:fire:` when removing code or files
    * :green_heart: `:green_heart:` when fixing the CI build
    * :white_check_mark: `:white_check_mark:` when adding tests
    * :lock: `:lock:` when dealing with security
    * :arrow_up: `:arrow_up:` when upgrading dependencies
    * :arrow_down: `:arrow_down:` when downgrading dependencies
    * :shirt: `:shirt:` when removing linter warnings

### Golang Style Guide

All Golang code is linted with [golangci-lint](https://github.com/golangci/golangci-lint).

* Avoid nesting by handling errors first
  ```go
  func (g *Gopher) WriteTo(w io.Writer) (size int64, err error) {
      err = binary.Write(w, binary.LittleEndian, int32(len(g.Name)))
      if err != nil {
          return
      }
      size += 4
      n, err := w.Write([]byte(g.Name))
      size += int64(n)
      if err != nil {
          return
      }
      err = binary.Write(w, binary.LittleEndian, int64(g.AgeYears))
      if err == nil {
          size += 4
      }
      return
  }
  ```
  Less nesting means less cognitive load on the reader
* Avoid repetition when possible
  
  Deploy one-off utility types for simpler code
  ```go
  type binWriter struct {
      w    io.Writer
      size int64
      err  error
  }
  // Write writes a value to the provided writer in little endian form.
  func (w *binWriter) Write(v interface{}) {
      if w.err != nil {
          return
      }
      if w.err = binary.Write(w.w, binary.LittleEndian, v); w.err == nil {
          w.size += int64(binary.Size(v))
      }
  }
  ```
* Document your code
    * Package name, with the associated documentation before.
      ```go
      // Package playground registers an HTTP handler at "/compile" that
      // proxies requests to the golang.org playground service.
      package playground
      ```
    * Exported identifiers appear in godoc, they should be documented correctly.
      ```go
      // Author represents the person who wrote and/or is presenting the document.
      type Author struct {
          Elem []Elem
      }
      
      // TextElem returns the first text elements of the author details.
      // This is used to display the author' name, job title, and company
      // without the contact details.
      func (p *Author) TextElem() (elems []Elem) {
      ```
* Shorter is better

  or at least _longer is not always better_.
  
  Try to find the shortest name that is self-explanatory.
    * Prefer MarshalIndent to MarshalWithIndentation.

  Don't forget that the package name will appear before the identifier you chose.
    * In package encoding/json we find the type Encoder, not JSONEncoder.
    * It is referred as json.Encoder.

### Specs Styleguide

TODO

#### Example

TODO

### Documentation Styleguide

* Use [godoc](https://blog.golang.org/godoc).

#### Example

```coffee
// Package sort provides primitives for sorting slices and user-defined
// collections.
package sort

// Fprint formats using the default formats for its operands and writes to w.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Fprint(w io.Writer, a ...interface{}) (n int, err error) {
```

## Additional Notes

### Issue and Pull Request Labels

This section lists the labels we use that helps us track and manage issues and pull requests. Most labels used are common across all itscontained repositories, but some are specific to `itscontained/secret-manager`.

[GitHub search](https://help.github.com/articles/searching-issues/) makes it easy to use labels for finding groups of issues or pull requests you're interested in. 

The labels loosely group by their purpose, but it's not required that every issue have a label from every group or that an issue can't have more than one label from the same group.

Please open an issue on `itscontained/secret-manager` if you have suggestions for new labels, and if you notice some labels are missing on some repositories, then please open an issue on that repository.

#### Type of Issue and Issue State

TODO

#### Topic Categories

TODO

[beginner]:https://github.com/search?q=is%3Aopen+is%3Aissue+label%3Abeginner+label%3Ahelp-wanted+user%3Aitscontained+sort%3Acomments-desc
[help-wanted]:https://github.com/search?q=is%3Aopen+is%3Aissue+label%3Ahelp-wanted+user%3Aitscontained+sort%3Acomments-desc+-label%3Abeginner
