import {Component, ViewChild, ElementRef, AfterViewInit} from '@angular/core';
import {NzCardComponent} from 'ng-zorro-antd/card';
import {NzSpinComponent} from 'ng-zorro-antd/spin';
import {TemplateBase} from '../template-base.component';
import {ValueContent} from '../template';
import {CommonModule} from '@angular/common';
import {isFunction} from 'rxjs/internal/util/isFunction';

@Component({
  selector: 'app-value',
  imports: [
    CommonModule,
    NzCardComponent,
    NzSpinComponent
],
  templateUrl: './value.component.html',
  standalone: true,
  styleUrl: './value.component.scss',
  //inputs: ['app', 'page', 'content', 'params', 'data', 'isChild']
})
export class ValueComponent extends TemplateBase implements AfterViewInit {
  @ViewChild("valueElement", {static: false}) valueElement!: ElementRef;
    
  override ngAfterViewInit() {
    super.ngAfterViewInit();
    this.updateDisplayValue();
  }
  
  // 数据加载后更新显示
  onDataLoaded() {
    this.updateDisplayValue();
  }
  
  // 更新显示数值
  updateDisplayValue() {
    const content = this.content as ValueContent;
    if (!content) return;
    
    // 获取新数值
    let newValue: number | string;
    if (content.value !== undefined) {
      newValue = content.value;
    } else if (this.data) {
      newValue = this.data || 0;
    } else {
      newValue = 0;
    }
    
    // 检查数值是否变化
    if (newValue !== this.data) {
      // 触发变化回调
      if (typeof content.onChange == "string" && content.onChange.length > 0) {
        try {
          content.onChange = new Function("oldValue", "newValue", content.onChange)
        } catch (e) {
          console.error(e)
        }
      }
      if (isFunction(content.onChange)) {
        content.onChange.call(this, this.data, newValue);
      }
      
      this.data = newValue;
    }
  }
  
  // 格式化数值显示
  formatValue(value: number | string): string {
    const content = this.content as ValueContent;
    if (!content) return String(value);
    
    if (typeof value === 'number') {
      // 小数位数
      const decimals = content.decimals || 0;
      let formatted = value.toFixed(decimals);
      
      // 格式化类型
      if (content.format === 'currency') {
        formatted = '¥' + formatted;
      } else if (content.format === 'percent') {
        formatted = formatted + '%';
      } else if (content.format === 'bytes') {
        formatted = this.formatBytes(value);
      } else if (content.format === 'thousands') {
        formatted = this.formatThousands(value);
      }
      
      return formatted;
    }
    
    return String(value);
  }
  
  // 格式化字节
  formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
  
  // 格式化千分位
  formatThousands(num: number): string {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  }
  
  // 获取样式
  getValueStyle(): any {
    const content = this.content as ValueContent;
    if (!content) return {};
    
    return {
      'font-size': content.fontSize || '48px',
      'color': content.color || '#1890ff',
      'font-weight': content.fontWeight || 'bold',
      'font-family': content.fontFamily || 'Arial, sans-serif',
      'text-align': content.align || 'center',
      'background': content.background || 'transparent',
    };
  }
  
  // 获取容器样式
  getContainerStyle(): any {
    const content = this.content as ValueContent;
    const baseStyle = content.style || {};
    const bodyStyle = content.bodyStyle || {};
    
    return {
      ...baseStyle,
      ...bodyStyle,
      'display': 'flex',
      'flex-direction': 'column',
      'justify-content': 'center',
      'align-items': content.align === 'left' ? 'flex-start' : 
                     content.align === 'right' ? 'flex-end' : 'center',
    };
  }
  
  // 获取图片样式
  getImageStyle(): any {
    const content = this.content as ValueContent;
    if (!content || !content.image) return {};
    
    const size = content.imageSize;
    let width = '48px';
    let height = '48px';
    
    if (size) {
      if (typeof size === 'number') {
        width = size + 'px';
        height = size + 'px';
      } else {
        width = size;
        height = size;
      }
    }
    
    return { width, height };
  }
  
  // 获取wrapper样式类
  getWrapperClass(): string {
    const content = this.content as ValueContent;
    if (!content) return '';
    
    // 当图片在top或bottom位置时，wrapper需要垂直排列
    if (content.imagePosition === 'top' || content.imagePosition === 'bottom') {
      return 'vertical';
    }
    return '';
  }
}
