<template>
  <router-view />
</template>

<script setup lang="ts">
// App root component
</script>

<style>
/* ======== Global Reset ======== */
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { background: #0A0F1C; color: #EDF0F5; }
#app { width: 100%; height: 100vh; overflow: hidden; }
body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Segoe UI', Roboto, sans-serif;
  -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;
}

/* ======== Element Plus Global Dark Theme ======== */

/* -- Card -- */
.el-card { background: rgba(255,255,255,0.025) !important; border: 1px solid rgba(255,255,255,0.06) !important; border-radius: 14px !important; color: #EDF0F5 !important; }
.el-card__header { border-bottom: 1px solid rgba(255,255,255,0.06) !important; color: #EDF0F5 !important; }

/* -- Table -- */
.el-table { background: transparent !important; color: #EDF0F5 !important; --el-table-tr-bg-color:transparent; --el-table-header-bg-color:rgba(255,255,255,0.03); --el-table-border-color:rgba(255,255,255,0.06); }
.el-table th { background: rgba(255,255,255,0.03) !important; color: #8890A5 !important; font-weight: 600 !important; border-color: rgba(255,255,255,0.06) !important; }
.el-table td { border-color: rgba(255,255,255,0.04) !important; }
.el-table tr { background: transparent !important; }
.el-table--striped .el-table__body tr.el-table__row--striped td { background: rgba(255,255,255,0.015) !important; }
.el-table__body tr:hover > td { background: rgba(255,255,255,0.04) !important; }

/* -- Pagination -- */
.el-pagination { --el-pagination-bg-color:transparent; --el-pagination-text-color:#8890A5; }
.el-pagination button { background: rgba(255,255,255,0.04) !important; color: #8890A5 !important; border: 1px solid rgba(255,255,255,0.06) !important; border-radius: 8px !important; }
.el-pagination button:hover { color: #00F5FF !important; border-color: #00F5FF !important; }
.el-pagination .el-pager li { background: rgba(255,255,255,0.04) !important; color: #8890A5 !important; border-radius: 8px !important; }
.el-pagination .el-pager li.active { background: #00F5FF !important; color: #0A0F1C !important; }

/* -- Buttons -- */
.el-button--primary { --el-button-bg-color:#00F5FF; --el-button-border-color:#00F5FF; --el-button-text-color:#0A0F1C; --el-button-hover-bg-color:#0DE0E8; --el-button-hover-border-color:#0DE0E8; }
.el-button--default { --el-button-bg-color:rgba(255,255,255,0.05); --el-button-border-color:rgba(255,255,255,0.1); --el-button-text-color:#EDF0F5; }
.el-button--danger { --el-button-bg-color:rgba(248,113,113,0.15); --el-button-border-color:rgba(248,113,113,0.25); --el-button-text-color:#F87171; }

/* ======== INPUTS - Core ======== */
/* el-input */
.el-input__wrapper {
  background: rgba(255,255,255,0.05) !important;
  border: 1px solid rgba(255,255,255,0.08) !important;
  box-shadow: none !important; border-radius: 10px !important;
  transition: all .2s !important;
}
.el-input__wrapper:hover { border-color: rgba(255,255,255,0.15) !important; }
.el-input__wrapper.is-focus, .el-input__wrapper:focus-within {
  border-color: #00F5FF !important;
  box-shadow: 0 0 0 3px rgba(0,245,255,0.08) !important;
}
.el-input__inner { color: #EDF0F5 !important; font-size: 14px !important; }
.el-input__inner::placeholder, .el-input__placeholder { color: #64748B !important; }
/* el-input with prefix/suffix */
.el-input-group__append, .el-input-group__prepend {
  background: rgba(255,255,255,0.04) !important; border-color: rgba(255,255,255,0.08) !important;
  color: #8890A5 !important;
}

/* -- Textarea -- */
.el-textarea__inner {
  background: rgba(255,255,255,0.05) !important;
  border: 1px solid rgba(255,255,255,0.08) !important;
  color: #EDF0F5 !important; border-radius: 10px !important;
  font-size: 14px !important; line-height: 1.6 !important;
  transition: all .2s !important; box-shadow: none !important;
}
.el-textarea__inner:hover { border-color: rgba(255,255,255,0.15) !important; }
.el-textarea__inner:focus {
  border-color: #00F5FF !important;
  box-shadow: 0 0 0 3px rgba(0,245,255,0.08) !important;
}
.el-textarea__inner::placeholder { color: #64748B !important; }
.el-textarea .el-input__count { background: transparent !important; color: #64748B !important; }

/* -- Select -- */
.el-select .el-input__wrapper { background: rgba(255,255,255,0.05) !important; }
.el-select-dropdown { background: #111827 !important; border: 1px solid rgba(255,255,255,0.06) !important; }
.el-select-dropdown__item { color: #8890A5 !important; }
.el-select-dropdown__item.hover, .el-select-dropdown__item:hover { background: rgba(0,245,255,0.08) !important; color: #00F5FF !important; }
.el-select-dropdown__item.selected { color: #00F5FF !important; font-weight: 600 !important; }
.el-popper__arrow::before { background: #111827 !important; border-color: rgba(255,255,255,0.06) !important; }

/* -- Tree Select -- */
.el-tree-select .el-input__wrapper { background: rgba(255,255,255,0.05) !important; }
.el-tree { background: transparent !important; color: #EDF0F5 !important; }
.el-tree-node__content:hover { background: rgba(0,245,255,0.06) !important; }
.el-tree-node.is-current > .el-tree-node__content { background: rgba(0,245,255,0.1) !important; color: #00F5FF !important; }

/* -- Input Number -- */
.el-input-number .el-input__wrapper { background: rgba(255,255,255,0.05) !important; }
.el-input-number__decrease, .el-input-number__increase {
  background: rgba(255,255,255,0.04) !important; color: #8890A5 !important;
  border-color: rgba(255,255,255,0.08) !important;
}
.el-input-number__decrease:hover, .el-input-number__increase:hover { color: #00F5FF !important; }

/* -- Date Picker -- */
.el-date-editor .el-input__wrapper { background: rgba(255,255,255,0.05) !important; }
.el-date-editor .el-input__inner { color: #EDF0F5 !important; }
.el-picker-panel { background: #111827 !important; border: 1px solid rgba(255,255,255,0.06) !important; color: #EDF0F5 !important; }
.el-date-picker__header-label { color: #EDF0F5 !important; }
.el-date-table td.available:hover { color: #00F5FF !important; }
.el-date-table td.current:not(.disabled) span { background: #00F5FF !important; color: #0A0F1C !important; }
.el-date-table td.today span { color: #00F5FF !important; }
.el-month-table td.current:not(.disabled) .cell, .el-year-table td.current:not(.disabled) .cell { background: #00F5FF !important; color: #0A0F1C !important; }
.el-picker-panel__icon-btn:hover { color: #00F5FF !important; }

/* -- Switch -- */
.el-switch__core { background: rgba(255,255,255,0.12) !important; border-color: transparent !important; }
.el-switch.is-checked .el-switch__core { background: #00F5FF !important; border-color: #00F5FF !important; }

/* -- Radio / Checkbox -- */
.el-radio__label, .el-checkbox__label { color: #EDF0F5 !important; }
.el-radio__inner { background: rgba(255,255,255,0.06) !important; border-color: rgba(255,255,255,0.15) !important; }
.el-radio.is-checked .el-radio__inner { background: #00F5FF !important; border-color: #00F5FF !important; }
.el-checkbox__inner { background: rgba(255,255,255,0.06) !important; border-color: rgba(255,255,255,0.15) !important; }
.el-checkbox.is-checked .el-checkbox__inner { background: #00F5FF !important; border-color: #00F5FF !important; }

/* -- Form Label -- */
.el-form-item__label { color: #8890A5 !important; font-weight: 500 !important; }

/* -- Dialog / Modal -- */
.el-dialog { background: #111827 !important; border: 1px solid rgba(255,255,255,0.06) !important; border-radius: 16px !important; }
.el-dialog__title { color: #EDF0F5 !important; font-weight: 700 !important; }
.el-dialog__headerbtn .el-dialog__close { color: #8890A5 !important; font-size: 18px !important; }
.el-dialog__headerbtn:hover .el-dialog__close { color: #F87171 !important; }
.el-overlay { background: rgba(0,0,0,0.55) !important; backdrop-filter: blur(6px); }

/* -- MessageBox -- */
.el-message-box { background: #111827 !important; border: 1px solid rgba(255,255,255,0.06) !important; }
.el-message-box__title { color: #EDF0F5 !important; }
.el-message-box__message { color: #8890A5 !important; }

/* -- Tags -- */
.el-tag { border-radius: 100px !important; }
.el-tag--success { background: rgba(16,185,129,0.12) !important; color: #10B981 !important; border-color: transparent !important; }
.el-tag--danger { background: rgba(248,113,113,0.12) !important; color: #F87171 !important; border-color: transparent !important; }
.el-tag--warning { background: rgba(245,158,11,0.12) !important; color: #F59E0B !important; border-color: transparent !important; }
.el-tag--info { background: rgba(255,255,255,0.06) !important; color: #8890A5 !important; border-color: transparent !important; }

/* -- Tabs -- */
.el-tabs__item { color: #8890A5 !important; }
.el-tabs__item.is-active { color: #00F5FF !important; }
.el-tabs__active-bar { background: #00F5FF !important; }
.el-tabs__nav-wrap::after { background: rgba(255,255,255,0.06) !important; }

/* -- Empty -- */
.el-empty__description p { color: #8890A5 !important; }

/* -- Upload -- */
.el-upload { background: transparent !important; }
.el-upload-dragger { background: rgba(255,255,255,0.03) !important; border: 1px dashed rgba(255,255,255,0.1) !important; }
.el-upload-dragger:hover { border-color: #00F5FF !important; }

/* -- Image Viewer -- */
.el-image-viewer__wrapper { background: rgba(0,0,0,0.92) !important; }

/* -- Loading -- */
.el-loading-mask { background: rgba(10,15,28,0.6) !important; backdrop-filter: blur(2px); }
</style>

