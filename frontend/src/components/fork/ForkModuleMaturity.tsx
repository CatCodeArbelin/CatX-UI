import type { ReactNode } from 'react';
import { CodeOutlined } from '@ant-design/icons';
import { Tag, Tooltip } from 'antd';
import { useTranslation } from 'react-i18next';

import {
  forkModuleRegistry,
  type ForkModuleDefinition,
  type ForkModuleId,
} from '@/forkext/registry';
import './ForkModuleMaturity.css';

interface ForkModuleMaturityProps {
  moduleId: ForkModuleId;
  showName?: boolean;
}

export function ForkModuleMaturity({ moduleId, showName = false }: ForkModuleMaturityProps) {
  const { t } = useTranslation();
  const definition: ForkModuleDefinition = forkModuleRegistry[moduleId];

  if (definition.maturity === 'ready') return null;

  const moduleName = t(definition.labelKey);
  const maturityLabel = definition.maturity.toUpperCase();
  const explanation = t('fork.maturity.explanation');
  const tooltip = t('fork.maturity.tooltip', { explanation });
  const accessibleLabel = t('fork.maturity.accessible', {
    module: moduleName,
    maturity: maturityLabel,
    explanation,
  });

  return (
    <span className={`fork-module-maturity${showName ? ' fork-module-maturity--named' : ''}`}>
      {showName && <span className="fork-module-maturity__name">{moduleName}</span>}
      <Tooltip title={tooltip} placement="top">
        <span className="fork-module-maturity__trigger" role="status" aria-label={accessibleLabel}>
          <Tag className="fork-maturity-badge" color="geekblue" icon={<CodeOutlined />}>
            {maturityLabel}
          </Tag>
        </span>
      </Tooltip>
    </span>
  );
}

export function ForkModuleTitle({
  moduleId,
  title,
  compact = false,
}: {
  moduleId: ForkModuleId;
  title: ReactNode;
  compact?: boolean;
}) {
  return (
    <span className={`fork-module-title${compact ? ' fork-module-title--compact' : ''}`}>
      <span
        className="fork-module-title__text"
        title={typeof title === 'string' ? title : undefined}
      >
        {title}
      </span>
      <ForkModuleMaturity moduleId={moduleId} />
    </span>
  );
}
