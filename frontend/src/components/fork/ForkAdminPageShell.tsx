import type { ReactNode } from 'react';
import { ConfigProvider, Layout } from 'antd';

import { useTheme } from '@/hooks/useTheme';
import AppSidebar from '@/layouts/AppSidebar';

interface ForkAdminPageShellProps {
  pageClass: string;
  children: ReactNode;
}

/**
 * CatX admin pages use the same composition as upstream panel pages without
 * moving the shared sidebar into PanelLayout (which would duplicate it).
 */
export default function ForkAdminPageShell({ pageClass, children }: ForkAdminPageShellProps) {
  const { isDark, isUltra, antdThemeConfig } = useTheme();
  const classes = ['fork-admin-page', pageClass, isDark ? 'is-dark' : '', isUltra ? 'is-ultra' : '']
    .filter(Boolean)
    .join(' ');

  return (
    <ConfigProvider theme={antdThemeConfig}>
      <Layout className={classes} data-testid="fork-admin-shell">
        <AppSidebar />
        <Layout className="content-shell">
          <Layout.Content id="content-layout" className="content-area">
            {children}
          </Layout.Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
}
