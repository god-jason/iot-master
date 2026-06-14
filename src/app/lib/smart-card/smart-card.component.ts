import {Component, Input} from '@angular/core';
import {CommonModule, NgOptimizedImage} from '@angular/common';

export type CardTheme = 'white' | 'dark' | 'transparent' | 'macaron' | 'dark-blue';

@Component({
  selector: 'smart-card',
  standalone: true,
  imports: [
    CommonModule,
  ],
  templateUrl: './smart-card.component.html',
  styleUrl: './smart-card.component.scss'
})
export class SmartCardComponent {
  @Input() title?: string;
  @Input() icon?: string; // 图片URL路径
  @Input() titleStyle?: { [key: string]: any };
  @Input() bodyStyle?: { [key: string]: any };
  @Input() theme: CardTheme = 'white';

  get themeClass(): string {
    return `smart-card--${this.theme}`;
  }

  get hasTitle(): boolean {
    return !!this.title;
  }
}
