# frozen_string_literal: true

require 'net/http'
require 'uri'
require 'json'

# VoiceVoxとやり取りするためのクラス
class VoiceVox
  def initialize(config)
    @host = config.voicevox.host
  end

  # speakにtextとspeakerさえ渡せば音声を生成してくれる
  def speak(text, speaker = 2)
    query = voice_query(text, speaker)
    generate_voice(query, speaker)
  end

  private

  # VoiceVoxのクエリを作成
  def voice_query(text, speaker)
    uri = URI.parse("#{@host}/audio_query")
    params = { speaker: speaker, text: text }
    uri.query = URI.encode_www_form(params)
    response = Net::HTTP.post_form(uri, {})
    puts "[VoiceVox-Query] Text: #{text}"
    response.body
  end

  # クエリから音声を生成
  def generate_voice(query, speaker)
    uri = URI.parse("#{@host}/synthesis")
    params = { speaker: speaker }
    uri.query = URI.encode_www_form(params)
    header = { 'Content-Type' => 'application/json' }
    response = Net::HTTP.post(uri, query, header)
    response.body
  end
end
