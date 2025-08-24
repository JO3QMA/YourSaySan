# frozen_string_literal: true

require 'net/http'
require 'uri'
require 'json'

# VoiceVoxとやり取りするためのクラス
class VoiceVox
  def initialize(config, logger)
    @host = config.voicevox.host
    @logger = logger
  end

  # textとspeakerから音声を生成
  def speak(text, speaker = 2)
    @logger.info('VoiceVox') do
      "Speak: Text: #{text} Speaker: #{speaker}"
    end
    query = voice_query(text, speaker)
    generate_voice(query, speaker)
  end

  # 利用可能な話者の一覧を取得
  def get_speakers
    response = get_req('/speakers')
    if response && response.code == '200'
      speakers_data = JSON.parse(response.body)
      speakers = {}

      speakers_data.each do |speaker|
        speaker_name = speaker['name']
        speaker['styles'].each do |style|
          style_id = style['id']
          style_name = style['name']
          speakers[style_id] = "#{speaker_name}（#{style_name}）"
        end
      end

      # 話者ID順にソート
      speakers = speakers.sort.to_h

      @logger.info('VoiceVox') do
        "Retrieved #{speakers.length} speaker styles from VoiceVox API"
      end
      speakers
    else
      @logger.error('VoiceVox') do
        "Failed to get speakers from VoiceVox API: #{response&.code}"
      end
      nil
    end
  rescue StandardError => e
    @logger.error('VoiceVox') do
      "Error getting speakers from VoiceVox API: #{e.message}"
    end
    nil
  end

  private

  # VoiceVoxのクエリを作成
  def voice_query(text, speaker)
    response = post_req('/audio_query', { speaker: speaker, text: text })
    @logger.debug('VoiceVox') do
      "Audio_query: Code: #{response.code} Text: #{text}"
    end
    response.body
  end

  # クエリから音声を生成
  def generate_voice(query, speaker)
    response = post_req('/synthesis', { speaker: speaker }, query)
    @logger.debug('VoiceVox') do
      "Synthesis: Code: #{response.code}"
    end
    response.body
  end

  # Get送信用関数
  def get_req(endpoint)
    uri = URI.join(@host, endpoint)
    http = Net::HTTP.new(uri.host, uri.port)
    http.open_timeout = 3
    http.read_timeout = 10
    request = Net::HTTP::Get.new(uri.request_uri)
    begin
      res = http.request(request)
      res.value
      res
    rescue Net::OpenTimeout, Net::ReadTimeout
      @logger.error('VoiceVox') do
        'VoiceVox request timed out.'
      end
      nil
    rescue StandardError => e
      @logger.error('VoiceVox') do
        e.message
      end
      nil
    end
  end

  # Post送信用関数
  def post_req(endpoint, query, data = '')
    uri = URI.join(@host, endpoint)
    uri.query = URI.encode_www_form(query)
    header = { 'Content-Type' => 'application/json' }
    http = Net::HTTP.new(uri.host, uri.port)
    http.open_timeout = 3
    http.read_timeout = 10
    request = Net::HTTP::Post.new(uri.request_uri, header)
    request.body = data
    begin
      res = http.request(request)
      res.value
      res
    rescue Net::OpenTimeout, Net::ReadTimeout
      @logger.error('VoiceVox') do
        'VoiceVox request timed out.'
      end
      nil
    rescue StandardError => e
      @logger.error('VoiceVox') do
        e.message
      end
      nil
    end
  end
end
