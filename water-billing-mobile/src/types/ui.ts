// Navigation item
export interface NavItem {
  title: string;
  href: string;
  icon?: string;
  roles?: string[]; // Which roles can see this
  children?: NavItem[];
}

// Sidebar props
export interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

// Chart data point
export interface ChartDataPoint {
  name: string;
  value: number;
  [key: string]: any;
}

// Table column definition
export interface TableColumn<T = any> {
  key: keyof T | string;
  title: string;
  render?: (value: any, record: T) => React.ReactNode;
  sortable?: boolean;
  width?: string | number;
}

// Filter option
export interface FilterOption {
  label: string;
  value: string | number;
}

// Dropdown option
export interface DropdownOption {
  label: string;
  value: string | number;
  disabled?: boolean;
}

// Toast notification
export interface Toast {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  title?: string;
  message: string;
  duration?: number;
}

// Mobile app config
export interface AppConfig {
  apiUrl: string;
  appName: string;
  appVersion: string;
  buildNumber?: string;
  environment: 'development' | 'staging' | 'production';
}