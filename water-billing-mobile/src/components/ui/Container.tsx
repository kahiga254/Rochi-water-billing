import { cn } from '@/lib/utils/cn';

interface ContainerProps {
  children: React.ReactNode;
  className?: string;
}

export function Container({ children, className }: ContainerProps) {
  return (
    <div className={cn('px-4 sm:px-6 lg:px-8 max-w-7xl mx-auto w-full', className)}>
      {children}
    </div>
  );
}