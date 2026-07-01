import {Component, ViewChild} from '@angular/core';
import { CommonModule } from '@angular/common';
import { TemplateBase } from '../template-base.component';
import { CodeContent} from '../template';
import { CodeEditor } from '@acrodata/code-editor';
import { LanguageSupport, LanguageDescription, StreamParser, StreamLanguage } from "@codemirror/language"
import { SmartToolbarComponent } from '../../lib/smart-toolbar/smart-toolbar.component';
import {NzSkeletonComponent} from 'ng-zorro-antd/skeleton';
import {isFunction} from 'rxjs/internal/util/isFunction';
import {FormsModule} from '@angular/forms';

function legacy(parser: StreamParser<unknown>): LanguageSupport {
  return new LanguageSupport(StreamLanguage.define(parser))
}
function sql(dialectName: keyof typeof import("@codemirror/lang-sql")) {
  return import("@codemirror/lang-sql").then(m => m.sql({dialect: (m as any)[dialectName]}))
}

@Component({
  selector: 'app-code',
  imports: [
    CommonModule,
    FormsModule,
    CodeEditor,
    NzSkeletonComponent,
    SmartToolbarComponent
  ],
  templateUrl: './code.component.html',
  styleUrl: './code.component.scss',
  standalone: true,
  inputs: ['page', 'content', 'params', 'data', 'isChild', 'pageComponent']
})
export class CodeComponent extends TemplateBase {

  @ViewChild("toolbar", {static: false}) toolbar!: SmartToolbarComponent;
  toolbarValue = {}

  override data: string = '';

  languages = [
    // New-style language modes
    LanguageDescription.of({
      name: "C",
      extensions: ["c","h","ino"],
      load() {
        return import("@codemirror/lang-cpp").then(m => m.cpp())
      }
    }),
    LanguageDescription.of({
      name: "C++",
      alias: ["cpp"],
      extensions: ["cpp","c++","cc","cxx","hpp","h++","hh","hxx"],
      load() {
        return import("@codemirror/lang-cpp").then(m => m.cpp())
      }
    }),
    LanguageDescription.of({
      name: "CSS",
      extensions: ["css"],
      load() {
        return import("@codemirror/lang-css").then(m => m.css())
      }
    }),
    LanguageDescription.of({
      name: "Go",
      extensions: ["go"],
      load() {
        return import("@codemirror/lang-go").then(m => m.go())
      }
    }),
    LanguageDescription.of({
      name: "HTML",
      alias: ["xhtml"],
      extensions: ["html", "htm", "handlebars", "hbs"],
      load() {
        return import("@codemirror/lang-html").then(m => m.html())
      }
    }),
    LanguageDescription.of({
      name: "Java",
      extensions: ["java"],
      load() {
        return import("@codemirror/lang-java").then(m => m.java())
      }
    }),
    LanguageDescription.of({
      name: "JavaScript",
      alias: ["ecmascript","js","node"],
      extensions: ["js", "mjs", "cjs"],
      load() {
        return import("@codemirror/lang-javascript").then(m => m.javascript())
      }
    }),
    LanguageDescription.of({
      name: "JSON",
      alias: ["json5"],
      extensions: ["json","map"],
      load() {
        return import("@codemirror/lang-json").then(m => m.json())
      }
    }),
    LanguageDescription.of({
      name: "Markdown",
      extensions: ["md", "markdown", "mkd"],
      load() {
        return import("@codemirror/lang-markdown").then(m => m.markdown())
      }
    }),
    LanguageDescription.of({
      name: "Python",
      extensions: ["BUILD","bzl","py","pyw"],
      filename: /^(BUCK|BUILD)$/,
      load() {
        return import("@codemirror/lang-python").then(m => m.python())
      }
    }),
    LanguageDescription.of({
      name: "SQL",
      extensions: ["sql"],
      load() { return sql("StandardSQL") }
    }),
    LanguageDescription.of({
      name: "XML",
      alias: ["rss","wsdl","xsd"],
      extensions: ["xml","xsl","xsd","svg"],
      load() {
        return import("@codemirror/lang-xml").then(m => m.xml())
      }
    }),
    LanguageDescription.of({
      name: "YAML",
      alias: ["yml"],
      extensions: ["yaml","yml"],
      load() {
        return import("@codemirror/lang-yaml").then(m => m.yaml())
      }
    }),
    LanguageDescription.of({
      name: "Lua",
      extensions: ["lua"],
      load() {
        return import("@codemirror/legacy-modes/mode/lua").then(m => legacy(m.lua))
      }
    }),
    LanguageDescription.of({
      name: "Properties files",
      alias: ["ini","properties"],
      extensions: ["properties","ini","in"],
      load() {
        return import("@codemirror/legacy-modes/mode/properties").then(m => legacy(m.properties))
      }
    }),
  ];

  onChange($event:any){
    const content = this.content as CodeContent
    if (!content) return

    if (typeof content.change == "string" && content.change.length > 0) {
      try {
        content.change = new Function("data", content.change)
      } catch (e) {
        console.error(e)
      }
    }
    if (isFunction(content.change)) {
      //this.submitting = true
      content.change.call(this, this.data)
    }
  }
}
