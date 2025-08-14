FROM ruby:3.4.2

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN bundle install --without development test

# Optional audio tooling; adjust as needed
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg opus-tools && rm -rf /var/lib/apt/lists/*

COPY . .

ENV RACK_ENV=production

CMD ["ruby", "run.rb"]


