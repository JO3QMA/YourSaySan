# frozen_string_literal: true

require 'net/http'
require 'uri'
require 'json'

# VoiceVoxとやり取りするためのクラス
class VoiceVox
  HOST = 'http://127.0.0.1:50021'

  def voice_query(text, speaker = 2)
    puts "[VoiceVox-Query] Text: #{text}"
    # リクエストURL
    uri = URI.parse("#{HOST}/audio_query")
    params = {
      speaker: speaker,
      text: text
    }
    uri.query = URI.encode_www_form(params)

    # HTTP POSTリクエストを作成
    response = Net::HTTP.post_form(uri, {})

    # レスポンスのステータスコードとボディを表示
    response.body
  end

  def speak(query, speaker = 2)
    uri = URI.parse("#{HOST}/synthesis")
    params = {
      speaker: speaker
    }
    uri.query = URI.encode_www_form(params)
    header = {
      'Content-Type' => 'application/json'
    }

    response = Net::HTTP.post(uri, query, header)

    puts "[VoiceVox-Speak] #{response.code}"
    response.body
  end
end
