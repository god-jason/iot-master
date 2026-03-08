import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ThemeService {

  theme = ''

  constructor() {
    let theme = localStorage.getItem("theme") || "dark"
    this.loadTheme(theme).then(console.log, console.error)
  }

  loadCss(href: string, id: string): Promise<Event> {
    return new Promise((resolve, reject) => {
      const style = document.createElement('link');
      style.rel = 'stylesheet';
      style.href = href;
      style.id = id;
      style.onload = resolve;
      style.onerror = reject;
      document.head.append(style);
    });
  }

  removeUnusedTheme(theme: string): void {
    document.documentElement.classList.remove(theme);
    const removedThemeStyle = document.getElementById(theme);
    if (removedThemeStyle) {
      document.head.removeChild(removedThemeStyle);
    }
  }

  loadTheme(theme: string): Promise<Event> {
    return new Promise((resolve, reject) => {
      this.loadCss(`${theme}.css`, theme).then(
        e => {
          document.documentElement.classList.add(theme);
          if(this.theme)this.removeUnusedTheme(this.theme);
          this.theme = theme
          resolve(e);
        },
        e => reject(e)
      );
    })
  }

  changeTheme(theme: string): void {
    if (this.theme == theme)
      return
    localStorage.setItem("theme", theme)
    this.loadTheme(theme).then(console.log, console.error)
  }
}
