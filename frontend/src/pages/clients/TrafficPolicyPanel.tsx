import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Descriptions,
  InputNumber,
  Space,
  Switch,
  Tag,
  Typography,
  message,
} from 'antd';
import { useTranslation } from 'react-i18next';

import { HttpUtil, SizeFormatter } from '@/utils';
import CatxState from '@/components/fork/CatxState';
import { ForkModuleTitle } from '@/components/fork/ForkModuleMaturity';
import type { ForkRuntimeState } from '@/lib/fork-feature';

interface TrafficPolicyView {
  enabled: boolean;
  windowSeconds: number;
  quotaBytes: number;
  activeUploadBps: number;
  activeDownloadBps: number;
  throttleUploadBps: number;
  throttleDownloadBps: number;
  lifecycle: string;
  reason: string;
  usedBytes: number;
  remainingBytes: number;
  enforcement: string;
  enforcementNote?: string;
  state?: ForkRuntimeState;
  featureDisabled?: boolean;
}

interface TrafficPolicyPanelProps {
  email: string;
  readOnly?: boolean;
}

export default function TrafficPolicyPanel({ email, readOnly = false }: TrafficPolicyPanelProps) {
  const { t } = useTranslation();
  const [view, setView] = useState<TrafficPolicyView | null>(null);
  const [loading, setLoading] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [featureState, setFeatureState] = useState<ForkRuntimeState>('loading');
  const [draft, setDraft] = useState<Partial<TrafficPolicyView>>({});
  const [messageApi, contextHolder] = message.useMessage();
  const reasonLabel = (reason: string) => {
    const key: Record<string, string> = {
      'window reset': 'fork.common.labels.windowResetReason',
      'fixed-window quota reached': 'fork.common.labels.quotaReachedReason',
      'upstream client disabled': 'fork.common.labels.upstreamDisabledReason',
    };
    return key[reason] ? t(key[reason]) : t('fork.common.unknown');
  };
  const enforcementNote = (note?: string) =>
    note === 'kernel attribution is not proven for generic Xray users'
      ? t('fork.common.labels.attributionUnsupportedReason')
      : note;
  const rate = (value: number) => `${SizeFormatter.sizeFormat(value)}/s`;

  const configure = () => {
    const next: TrafficPolicyView = {
      enabled: true,
      windowSeconds: 3600,
      quotaBytes: 0,
      activeUploadBps: 0,
      activeDownloadBps: 0,
      throttleUploadBps: 0,
      throttleDownloadBps: 0,
      lifecycle: 'active',
      reason: '',
      usedBytes: 0,
      remainingBytes: 0,
      enforcement: 'unsupported',
      state: 'active',
      featureDisabled: false,
    };
    setView(next);
    setDraft(next);
    setFeatureState('active');
  };

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const response = (await HttpUtil.get(
        `/panel/api/traffic-control/clients/${encodeURIComponent(email)}/policy`,
        undefined,
        { silent: true },
      )) as {
        success?: boolean;
        obj?: TrafficPolicyView;
        msg?: string;
        featureDisabled?: boolean;
      };
      if (!response.success) {
        setView(null);
        setFeatureState(
          response.featureDisabled || response.obj?.featureDisabled ? 'feature_off' : 'error',
        );
        return;
      }
      const next = response.obj ?? null;
      if (next?.state === 'unconfigured') {
        setView(null);
        setFeatureState('unconfigured');
        return;
      }
      setView(next);
      setFeatureState(next?.state || 'active');
      if (next) setDraft(next);
    } finally {
      setLoading(false);
    }
  }, [email]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function reset() {
    setResetting(true);
    try {
      const response = (await HttpUtil.post(
        `/panel/api/traffic-control/clients/${encodeURIComponent(email)}/policy/reset`,
      )) as { success?: boolean; obj?: TrafficPolicyView; msg?: string };
      if (!response.success) {
        messageApi.error(t('fork.common.labels.actionFailed'));
        return;
      }
      setView(response.obj ?? null);
      setFeatureState(response.obj?.state || 'active');
    } finally {
      setResetting(false);
    }
  }

  async function save() {
    setSaving(true);
    try {
      const response = (await HttpUtil.put(
        `/panel/api/traffic-control/clients/${encodeURIComponent(email)}/policy`,
        {
          ...draft,
          // Generic Xray-user rate shaping has no proven production
          // attribution provider. Never submit values that imply it is
          // active from the normal editor.
          ...(view && ['unsupported', 'degraded', 'disabled'].includes(view.enforcement)
            ? {
                activeUploadBps: 0,
                activeDownloadBps: 0,
                throttleUploadBps: 0,
                throttleDownloadBps: 0,
              }
            : {}),
        },
        { headers: { 'Content-Type': 'application/json' } },
      )) as { success?: boolean; obj?: TrafficPolicyView; msg?: string };
      if (!response.success) {
        messageApi.error(t('fork.common.labels.actionFailed'));
        return;
      }
      setView(response.obj ?? null);
      if (response.obj) setDraft(response.obj);
    } finally {
      setSaving(false);
    }
  }

  if (!view) {
    if (featureState === 'loading') return null;
    return (
      <div style={{ marginTop: 16 }}>
        <CatxState
          state={featureState}
          feature="traffic_control"
          description={
            featureState === 'unconfigured'
              ? t('fork.common.labels.trafficControlUnconfigured')
              : featureState === 'error'
                ? t('fork.common.labels.featureUnavailable')
                : undefined
          }
          action={
            !readOnly && featureState === 'unconfigured' ? (
              <Button type="primary" size="small" onClick={configure}>
                {t('fork.common.labels.configureTrafficControl')}
              </Button>
            ) : undefined
          }
        />
      </div>
    );
  }
  const enforcementBlocked = ['unsupported', 'degraded', 'disabled'].includes(view.enforcement);
  const tagColor =
    view.enforcement === 'unsupported'
      ? 'orange'
      : view.enforcement === 'degraded'
        ? 'red'
        : 'green';
  return (
    <>
      {contextHolder}
      <Card
        size="small"
        title={<ForkModuleTitle moduleId="M05" title={t('fork.common.labels.trafficControl')} />}
        loading={loading}
        style={{ marginTop: 16 }}
        extra={
          !readOnly && (
            <Button size="small" type="primary" onClick={() => void save()} loading={saving}>
              {t('save')}
            </Button>
          )
        }
      >
        <Descriptions size="small" column={1}>
          <Descriptions.Item label={t('enabled')}>
            {readOnly ? (
              <Tag color={view.enabled ? 'green' : 'default'}>
                {view.enabled ? t('fork.common.enabled') : t('fork.common.disabled')}
              </Tag>
            ) : (
              <Switch
                checked={Boolean(draft.enabled)}
                onChange={(enabled) => setDraft((v) => ({ ...v, enabled }))}
              />
            )}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.quotaWindow')}>
            {readOnly ? (
              <Typography.Text>
                {draft.windowSeconds} {t('fork.common.labels.seconds')} ·{' '}
                {draft.quotaBytes ? SizeFormatter.sizeFormat(draft.quotaBytes) : '∞'}
              </Typography.Text>
            ) : (
              <Space wrap>
                <InputNumber
                  min={60}
                  value={draft.windowSeconds}
                  onChange={(windowSeconds) =>
                    setDraft((v) => ({ ...v, windowSeconds: windowSeconds ?? 3600 }))
                  }
                  addonAfter={t('fork.common.labels.seconds')}
                />
                <InputNumber
                  min={0}
                  value={draft.quotaBytes}
                  onChange={(quotaBytes) =>
                    setDraft((v) => ({ ...v, quotaBytes: quotaBytes ?? 0 }))
                  }
                  addonAfter={t('fork.common.labels.bytes')}
                />
              </Space>
            )}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.activeSpeed')}>
            {enforcementBlocked ? (
              <Typography.Text type="secondary">
                {t('fork.common.labels.rateUnsupported')}
              </Typography.Text>
            ) : readOnly ? (
              <Space wrap>
                <Typography.Text>
                  {t('fork.common.labels.uploadSpeed')}: {rate(view.activeUploadBps)}
                </Typography.Text>
                <Typography.Text>
                  {t('fork.common.labels.downloadSpeed')}: {rate(view.activeDownloadBps)}
                </Typography.Text>
              </Space>
            ) : (
              <Space wrap>
                <InputNumber
                  min={0}
                  value={draft.activeUploadBps}
                  onChange={(activeUploadBps) =>
                    setDraft((v) => ({ ...v, activeUploadBps: activeUploadBps ?? 0 }))
                  }
                  disabled={enforcementBlocked}
                  addonAfter={t('fork.common.labels.uploadSpeed')}
                />
                <InputNumber
                  min={0}
                  value={draft.activeDownloadBps}
                  onChange={(activeDownloadBps) =>
                    setDraft((v) => ({ ...v, activeDownloadBps: activeDownloadBps ?? 0 }))
                  }
                  disabled={enforcementBlocked}
                  addonAfter={t('fork.common.labels.downloadSpeed')}
                />
              </Space>
            )}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.throttleSpeed')}>
            {enforcementBlocked ? (
              <Typography.Text type="secondary">
                {t('fork.common.labels.rateUnsupported')}
              </Typography.Text>
            ) : readOnly ? (
              <Space wrap>
                <Typography.Text>
                  {t('fork.common.labels.uploadSpeed')}: {rate(view.throttleUploadBps)}
                </Typography.Text>
                <Typography.Text>
                  {t('fork.common.labels.downloadSpeed')}: {rate(view.throttleDownloadBps)}
                </Typography.Text>
              </Space>
            ) : (
              <Space wrap>
                <InputNumber
                  min={0}
                  value={draft.throttleUploadBps}
                  onChange={(throttleUploadBps) =>
                    setDraft((v) => ({ ...v, throttleUploadBps: throttleUploadBps ?? 0 }))
                  }
                  disabled={enforcementBlocked}
                  addonAfter={t('fork.common.labels.uploadSpeed')}
                />
                <InputNumber
                  min={0}
                  value={draft.throttleDownloadBps}
                  onChange={(throttleDownloadBps) =>
                    setDraft((v) => ({ ...v, throttleDownloadBps: throttleDownloadBps ?? 0 }))
                  }
                  disabled={enforcementBlocked}
                  addonAfter={t('fork.common.labels.downloadSpeed')}
                />
              </Space>
            )}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.lifecycle')}>
            <Tag
              color={
                view.lifecycle === 'throttled'
                  ? 'orange'
                  : view.lifecycle === 'disabled'
                    ? 'red'
                    : 'green'
              }
            >
              {t(`fork.common.states.${view.lifecycle}`, { defaultValue: view.lifecycle })}
            </Tag>
            {view.reason ? <span className="hint">{reasonLabel(view.reason)}</span> : null}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.enforcement')}>
            <Tag color={tagColor}>
              {t(`fork.common.states.${view.enforcement}`, { defaultValue: view.enforcement })}
            </Tag>
            {view.enforcementNote ? (
              <span className="hint">{enforcementNote(view.enforcementNote)}</span>
            ) : null}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.used')}>
            {SizeFormatter.sizeFormat(view.usedBytes)} /{' '}
            {view.quotaBytes > 0 ? SizeFormatter.sizeFormat(view.quotaBytes) : '∞'}
          </Descriptions.Item>
          <Descriptions.Item label={t('fork.common.labels.remaining')}>
            {view.quotaBytes > 0 ? SizeFormatter.sizeFormat(view.remainingBytes) : '∞'}
          </Descriptions.Item>
        </Descriptions>
        {!readOnly && (
          <Button
            size="small"
            onClick={() => void reset()}
            loading={resetting}
            disabled={view.lifecycle !== 'throttled'}
          >
            {t('reset')}
          </Button>
        )}
      </Card>
    </>
  );
}
