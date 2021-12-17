import InfoTooltip from '@components/InfoTooltip/InfoTooltip';
import Tooltip from '@components/Tooltip/Tooltip';
import { WebVitalDescriptor } from '@pages/Player/StreamElement/Renderers/WebVitals/utils/WebVitalsUtils';
import classNames from 'classnames';
import { motion } from 'framer-motion';
import React from 'react';

import styles from './Metric.module.scss';

interface Props {
    configuration: WebVitalDescriptor;
    value: number;
    name: string;
}

const SimpleMetric = ({ configuration, value, name }: Props) => {
    const valueScore = getValueScore(value, configuration);

    return (
        <div
            className={classNames(styles.simpleMetric, styles.metric, {
                [styles.goodScore]: valueScore === ValueScore.Good,
                [styles.needsImprovementScore]:
                    valueScore === ValueScore.NeedsImprovement,
                [styles.poorScore]: valueScore === ValueScore.Poor,
            })}
        >
            <span className={styles.name}>{name}</span>
        </div>
    );
};

export const DetailedMetric = ({ configuration, value, name }: Props) => {
    const valueScore = getValueScore(value, configuration);

    return (
        <div
            className={classNames(styles.metric, styles.detailedMetric, {
                [styles.goodScore]: valueScore === ValueScore.Good,
                [styles.needsImprovementScore]:
                    valueScore === ValueScore.NeedsImprovement,
                [styles.poorScore]: valueScore === ValueScore.Poor,
            })}
        >
            <span className={styles.name}>{name}</span>
            <ScoreVisualization value={value} configuration={configuration} />
        </div>
    );
};

export default SimpleMetric;

enum ValueScore {
    Good,
    NeedsImprovement,
    Poor,
}

function getValueScore(
    value: number,
    { maxGoodValue, maxNeedsImprovementValue }: WebVitalDescriptor
): ValueScore {
    if (value <= maxGoodValue) {
        return ValueScore.Good;
    }
    if (value <= maxNeedsImprovementValue) {
        return ValueScore.NeedsImprovement;
    }

    return ValueScore.Poor;
}

function getInfoTooltipText(
    configuration: WebVitalDescriptor,
    value: number
): React.ReactNode {
    const valueScore = getValueScore(value, configuration);

    let message = '';
    switch (valueScore) {
        case ValueScore.Poor:
            message = `Looks like you're not doing so hot for ${configuration.name} on this session.`;
            break;
        case ValueScore.NeedsImprovement:
            message = `You're scoring okay for ${configuration.name} on this session. You can do better though!`;
            break;
        case ValueScore.Good:
            message = `You're scoring AMAZINGLY for ${configuration.name} on this session!`;
            break;
    }

    return (
        <div
            // This is to prevent the stream element from collapsing from clicking on a link.
            onClick={(e) => {
                e.stopPropagation();
            }}
        >
            {message}{' '}
            <a
                href={configuration.helpArticle}
                target="_blank"
                rel="noreferrer"
            >
                Learn more about optimizing {configuration.name}.
            </a>
        </div>
    );
}

interface ScoreVisualizationProps {
    value: number;
    configuration: WebVitalDescriptor;
}

const ScoreVisualization = ({
    configuration,
    value,
}: ScoreVisualizationProps) => {
    const valueScore = getValueScore(value, configuration);
    const scorePosition = getScorePosition(configuration, value);
    let gapSpacing = 0;

    switch (valueScore) {
        case ValueScore.NeedsImprovement:
            gapSpacing = 2 * 1;
            break;
        case ValueScore.Poor:
            gapSpacing = 2 * 2;
            break;
    }

    return (
        <div className={styles.scoreVisualization}>
            <Tooltip
                title={() => getTooltipText(configuration, value)}
                mouseEnterDelay={0}
            >
                <motion.div
                    className={styles.scoreIndicator}
                    animate={{
                        left: `calc(${
                            scorePosition * 100
                        }% - calc(var(--size) / 2) + ${gapSpacing}px)`,
                    }}
                    transition={{
                        type: 'spring',
                    }}
                >
                    <span
                        className={classNames(styles.value, {
                            [styles.mirror]: valueScore === ValueScore.Poor,
                        })}
                    >
                        {value.toFixed(2)}
                        <span className={styles.units}>
                            {configuration.units}
                        </span>
                        <InfoTooltip
                            className={styles.infoTooltip}
                            title={getInfoTooltipText(configuration, value)}
                            align={{ offset: [-13, 0] }}
                            placement="topLeft"
                        />
                    </span>
                </motion.div>
            </Tooltip>
            <div
                className={classNames(styles.good, {
                    [styles.active]: valueScore === ValueScore.Good,
                })}
            ></div>
            <div
                className={classNames(styles.needsImprovement, {
                    [styles.active]: valueScore === ValueScore.NeedsImprovement,
                })}
            ></div>
            <div
                className={classNames(styles.poor, {
                    [styles.active]: valueScore === ValueScore.Poor,
                })}
            ></div>
        </div>
    );
};

const getScorePosition = (configuration: WebVitalDescriptor, value: number) => {
    const valueScore = getValueScore(value, configuration);
    let offset = 0;
    let min = 0;
    let max = 0;
    const OFFSET_AMOUNT = 0.33;

    switch (valueScore) {
        case ValueScore.Good:
            offset = OFFSET_AMOUNT * 0;
            min = 0;
            max = configuration.maxGoodValue;
            break;
        case ValueScore.NeedsImprovement:
            offset = OFFSET_AMOUNT * 1;
            min = configuration.maxGoodValue;
            max = configuration.maxNeedsImprovementValue;
            break;
        case ValueScore.Poor:
            offset = OFFSET_AMOUNT * 2;
            min = configuration.maxNeedsImprovementValue;
            max = Infinity;
            break;
    }

    // There's no upper value for a poor value so we generate a random value.
    if (max === Infinity) {
        return offset + Math.random() * OFFSET_AMOUNT;
    }

    const range = max - min;
    const percent = (value - min) / range;
    const relativePercent = OFFSET_AMOUNT * percent;

    return offset + relativePercent;
};

function getTooltipText(
    configuration: WebVitalDescriptor,
    value: number
): React.ReactNode {
    const message = `This session scored ${value.toFixed(2)} ${
        configuration.units
    }. An okay score is less than ${configuration.maxNeedsImprovementValue} ${
        configuration.units
    } and a great score is less than ${configuration.maxGoodValue} ${
        configuration.units
    }.`;

    return (
        <div
            // This is to prevent the stream element from collapsing from clicking on a link.
            onClick={(e) => {
                e.stopPropagation();
            }}
        >
            {message}{' '}
            <a
                href={configuration.helpArticle}
                target="_blank"
                rel="noreferrer"
            >
                Learn more about optimizing {configuration.name}.
            </a>
        </div>
    );
}
