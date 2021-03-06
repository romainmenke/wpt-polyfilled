package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestMain(t *testing.T) {
	srv := httptest.NewUnstartedServer(nil)
	srvAddr := srv.URL

	srv1 := httptest.NewUnstartedServer(nil)
	srv1Addr := srv1.URL

	srv2 := httptest.NewUnstartedServer(nil)
	srv2Addr := srv2.URL

	srv3 := httptest.NewUnstartedServer(nil)
	srv3Addr := srv3.URL

	srv.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}
	srv1.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}
	srv2.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}
	srv2.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}

	srv.Start()
	srv1.Start()
	srv2.Start()
	srv3.Start()

	defer srv.Close()
	defer srv1.Close()
	defer srv2.Close()
	defer srv3.Close()

	_, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJS(t *testing.T) {
	srv := httptest.NewUnstartedServer(nil)
	srvAddr := srv.URL

	srv1 := httptest.NewUnstartedServer(nil)
	srv1Addr := srv1.URL

	srv2 := httptest.NewUnstartedServer(nil)
	srv2Addr := srv2.URL

	srv3 := httptest.NewUnstartedServer(nil)
	srv3Addr := srv3.URL

	srv.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}
	srv1.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}
	srv2.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}
	srv2.Config = &http.Server{Handler: wptHandler(srvAddr, srv1Addr, srv2Addr, srv3Addr)}

	srv.Start()
	srv1.Start()
	srv2.Start()
	srv3.Start()

	defer srv.Close()
	defer srv1.Close()
	defer srv2.Close()
	defer srv3.Close()

	_, err := http.Get(srv.URL + "/resources/testharness.js")
	if err != nil {
		t.Fatal(err)
	}
}

func TestJSInHTML(t *testing.T) {
	source := []byte(`<!DOCTYPE html>
<meta charset=utf-8>
<title>Animation constructor</title>
<link rel="help" href="https://drafts.csswg.org/web-animations/#dom-animation-animation">
<script src="/resources/testharness.js"></script>
<script src="/resources/testharnessreport.js"></script>
<script src="../../testcommon.js"></script>
<body>
<div id="log"></div>
<div id="target"></div>
<script>
'use strict';

const gTarget = document.getElementById('target');

function createEffect() {
  return new KeyframeEffect(gTarget, { opacity: [0, 1] });
}

function createNull() {
  return null;
}

const gTestArguments = [
  {
    createEffect: createNull,
    timeline: null,
    expectedTimeline: null,
    expectedTimelineDescription: 'null',
    description: 'with null effect and null timeline'
  },
  {
    createEffect: createNull,
    timeline: document.timeline,
    expectedTimeline: document.timeline,
    expectedTimelineDescription: 'document.timeline',
    description: 'with null effect and non-null timeline'
  },
  {
    createEffect: createNull,
    expectedTimeline: document.timeline,
    expectedTimelineDescription: 'document.timeline',
    description: 'with null effect and no timeline parameter'
  },
  {
    createEffect: createEffect,
    timeline: null,
    expectedTimeline: null,
    expectedTimelineDescription: 'null',
    description: 'with non-null effect and null timeline'
  },
  {
    createEffect: createEffect,
    timeline: document.timeline,
    expectedTimeline: document.timeline,
    expectedTimelineDescription: 'document.timeline',
    description: 'with non-null effect and non-null timeline'
  },
  {
    createEffect: createEffect,
    expectedTimeline: document.timeline,
    expectedTimelineDescription: 'document.timeline',
    description: 'with non-null effect and no timeline parameter'
  },
];

for (const args of gTestArguments) {
  test(t => {
    const effect = args.createEffect();
    const animation = new Animation(effect, args.timeline);

    assert_not_equals(animation, null,
                      'An animation sohuld be created');
    assert_equals(animation.effect, effect,
                  'Animation returns the same effect passed to ' +
                  'the Constructor');
    assert_equals(animation.timeline, args.expectedTimeline,
                  'Animation timeline should be ' + args.expectedTimelineDescription);
    assert_equals(animation.playState, 'idle',
                  'Animation.playState should be initially \'idle\'');
  }, 'Animation can be constructed ' + args.description);
}

test(t => {
  const effect = new KeyframeEffect(null,
                                    { left: ['10px', '20px'] },
                                    { duration: 10000, fill: 'forwards' });
  const anim = new Animation(effect, document.timeline);
  anim.pause();
  assert_equals(effect.getComputedTiming().progress, 0.0);
  anim.currentTime += 5000;
  assert_equals(effect.getComputedTiming().progress, 0.5);
  anim.finish();
  assert_equals(effect.getComputedTiming().progress, 1.0);
}, 'Animation constructed by an effect with null target runs normally');

async_test(t => {
  const iframe = document.createElement('iframe');

  iframe.addEventListener('load', t.step_func(() => {
    const div = createDiv(t, iframe.contentDocument);
    const effect = new KeyframeEffect(div, null, 10000);
    const anim = new Animation(effect);
    assert_equals(anim.timeline, document.timeline);
    iframe.remove();
    t.done();
  }));

  document.body.appendChild(iframe);
}, 'Animation constructed with a keyframe that target element is in iframe');

</script>
</body>
`)

	re := regexp.MustCompile(`(?s)<\s*script[^>]*>([^<]+?)</script>`)

	submatchall := re.FindAllSubmatch(source, -1)
	for _, element := range submatchall {
		log.Println(string(element[0]))
		log.Println(string(element[1]))
	}
}
