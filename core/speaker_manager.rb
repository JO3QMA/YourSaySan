# frozen_string_literal: true

require 'redis'

# 話者設定を管理するクラス
class SpeakerManager
  def initialize(config, logger, voicevox)
    @redis = Redis.new(
      host: config.redis.host,
      port: config.redis.port,
      db: config.redis.db
    )
    @logger = logger
    @default_speaker = 2
    @voicevox = voicevox
  end

  # ユーザーの話者設定を取得
  def get_speaker(user_id)
    speaker = @redis.get("speaker:#{user_id}")
    if speaker
      speaker.to_i
    else
      @default_speaker
    end
  end

  # ユーザーの話者設定を保存
  def set_speaker(user_id, speaker_id)
    @redis.set("speaker:#{user_id}", speaker_id)
    @logger.info('SpeakerManager') { "Set speaker for user #{user_id}: #{speaker_id}" }
  end

  # VoiceVoxのAPIから利用可能な話者の一覧を取得
  def get_available_speakers
    speakers = @voicevox.get_speakers
    if speakers
      speakers
    else
      @logger.warn('SpeakerManager') { "Using fallback speakers list" }
      get_fallback_speakers
    end
  end

  # 話者IDが有効かチェック
  def valid_speaker?(speaker_id)
    speakers = get_available_speakers
    speakers.key?(speaker_id.to_i)
  end

  private

  # フォールバック用の話者一覧（APIが利用できない場合）
  def get_fallback_speakers
    {
      0 => '四国めたん（あまあま）',
      1 => 'ずんだもん（あまあま）',
      2 => '四国めたん（ノーマル）',
      3 => 'ずんだもん（ノーマル）',
      4 => '四国めたん（セクシー）',
      5 => 'ずんだもん（セクシー）',
      6 => '四国めたん（ツンツン）',
      7 => 'ずんだもん（ツンツン）',
      8 => '四国めたん（ささやき）',
      9 => 'ずんだもん（ささやき）',
      10 => '四国めたん（ヒソヒソ）',
      11 => 'ずんだもん（ヒソヒソ）',
      12 => '春日部つむぎ（ノーマル）',
      13 => '雨晴はう（ノーマル）',
      14 => '波音リツ（ノーマル）',
      15 => '玄野武宏（ノーマル）',
      16 => '白上虎太郎（ノーマル）',
      17 => '青山龍星（ノーマル）',
      18 => '冥鳴ひまり（ノーマル）',
      19 => '九州そら（あまあま）',
      20 => '九州そら（ノーマル）',
      21 => '九州そら（セクシー）',
      22 => '九州そら（ツンツン）',
      23 => '九州そら（ささやき）',
      24 => 'もち子さん（ノーマル）',
      25 => '剣崎雌雄（ノーマル）',
      26 => 'WhiteCUL（ノーマル）',
      27 => 'WhiteCUL（たのしい）',
      28 => 'WhiteCUL（かなしい）',
      29 => 'WhiteCUL（びえーん）',
      30 => '後鬼（人間ver.）',
      31 => '後鬼（ぬいぐるみver.）',
      32 => 'No.7（ノーマル）',
      33 => 'No.7（アナウンス）',
      34 => 'No.7（読み聞かせ）',
      35 => '白癒（ノーマル）',
      36 => '白癒（セクシー）',
      37 => '白癒（ツンツン）',
      38 => '白癒（ささやき）',
      39 => '白癒（ヒソヒソ）',
      40 => '白癒（わーい）',
      41 => '白癒（びくびく）',
      42 => '白癒（おこ）',
      43 => '白癒（みょーん）',
      44 => '白癒（すわー）',
      45 => '白癒（あ）',
      46 => '白癒（えー）',
      47 => '白癒（うー）',
      48 => '白癒（あー）',
      49 => '白癒（きょー）',
      50 => '白癒（きょー）'
    }
  end
end
